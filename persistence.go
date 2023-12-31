package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func (r Repository) connectToDB() (*sql.DB, error) {
	db, err := sql.Open(PostgresDriver, DataSourceName)
	if err != nil {
		fmt.Println("Cannot connect to postgres database.")
		return nil, errors.New(
			fmt.Sprintf("Cannot coonect to database: %s", err.Error()),
		)
	}
	return db, nil
}

func (r Repository) createNewPerson(newPerson Person) error {
	sqlStatement := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3, $4, $5)",
		TableName)
	insert, err := r.db.Prepare(sqlStatement)
	if err != nil {
		fmt.Println("Cannot prepare sql statement.")
		return errors.New(
			fmt.Sprintf("Cannot prepare sql statement: %s", err.Error()),
		)
	}
	result, err := insert.Exec(
		newPerson.Id, newPerson.Name, newPerson.Nick,
		newPerson.Birth, pq.Array(newPerson.Stack),
	)
	if err != nil {
    if pgerr, ok := err.(*pq.Error); ok {
      const constraintViolatedCode pq.ErrorCode = "23505"
      if pgerr.Code == constraintViolatedCode {
        return errors.New("Nick has alredy been taken")
      }
    }
		fmt.Println("Cannot exec sql query")
		return errors.New(
			fmt.Sprintf("Cannot create user. Please check the provided values."),
		)
	}
	affect, _ := result.RowsAffected()
	fmt.Println("Rows affected: ", affect)
	return nil
}

func (r Repository) convertRowToPerson(rows *sql.Rows) ([]Person, error) {
  var person Person
  var personRows []Person

  defer rows.Close()

  for rows.Next() {
    err := rows.Scan(
      &person.Id, &person.Name,
      &person.Nick, &person.Birth, pq.Array(&person.Stack),
    )
    if err != nil {
      return nil, errors.New(
        fmt.Sprintf("Cannot scan row to a Person struct: %s", err.Error()),
      )
    }
    personRows = append(personRows, person)
  }

  return personRows, nil
}

func (r Repository) findAll() ([]Person, error){
  query := fmt.Sprintf(
    "SELECT id, name, nick, birth_date, stack FROM %s", TableName,
  )
  rows, err := r.db.Query(query)
  if err != nil {
    return nil, errors.New(
      fmt.Sprintf("Cannot query postgres database: %s", err.Error()),
    )
  }
  var personRows []Person
  personRows, err = r.convertRowToPerson(rows)
  if err != nil {
    return nil, err
  }
  return personRows, nil
}

func (r Repository) findPersonById(id uuid.UUID) ([]Person, error) {
  query := fmt.Sprintf(
    "SELECT id, name, nick, birth_date, stack from %s WHERE id = '%s'",
    TableName, id.String(),
  )
  rows, err := r.db.Query(query)
  if err != nil {
    return nil, errors.New(
      fmt.Sprintf("Cannot find person by his ID: %s", err.Error()),
    )
  }

  var personOutOfTheRow []Person
  personOutOfTheRow, err = r.convertRowToPerson(rows)
  if err != nil {
    return nil, err
  }

  if len(personOutOfTheRow) == 0 {
    return nil, errors.New("There are no people with the provided ID")
  }

  return personOutOfTheRow, nil
}

func (r Repository) findPersonByPattern(pattern string) ([]Person, error) {
  query := fmt.Sprintf(
    `
    SELECT id, name, nick, birth_date, stack
    FROM %s
    WHERE to_tsquery('people', $1) @@ search
    LIMIT 50
    `, TableName,
  )
  rows, err := r.db.Query(query, pattern)
  if err != nil {
    return nil, errors.New(
      fmt.Sprintf("Cannot find person by pattern: %s", err.Error()),
    )
  }

  var personOutOfTheRow []Person
  personOutOfTheRow, err = r.convertRowToPerson(rows)
  if err != nil {
    return nil, err
  }

  return personOutOfTheRow, nil
}

func (r Repository) getPeopleCount() (int, error) {
  query := fmt.Sprintf("SELECT count(*) FROM %s", TableName)
  rows, err := r.db.Query(query)
  if err != nil {
    return 0, errors.New(
      fmt.Sprintf("Cannot return people count: %s", err.Error()),
    )
  }
  defer rows.Close()

  var peopleCount int
  for rows.Next() {
    err := rows.Scan(&peopleCount)
    if err != nil {
      return 0, errors.New(
        fmt.Sprintf("Cannot scan row: %s", err.Error()),
      )
    }
  }

	return peopleCount, nil
}

func (r Repository) testDB() error {
	newPerson := Person{
		Name:  "Bar",
		Nick:  "Bar bar",
		Birth: "2003-07-21",
		Stack: []string{"go"},
	}
	newPerson.Id, _ = uuid.NewV7()

	err := r.createNewPerson(newPerson)
	if err != nil {
		return err
	}

	return nil
}
