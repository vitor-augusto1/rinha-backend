package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type PeopleCountResponse struct {
  Count int `json:"count"`
}

type GetPersonResponse struct {
  PersonReturned Person `json:"person"`
}

type PersonCreatedResponse struct {
  Message string `json:"success"`
}

type Person struct {
  Id    uuid.UUID
  Nick  string    `json:"apelido"`
  Name  string    `json:"nome"`
  Birth string    `json:"nascimento"`
  Stack []string  `json:"stack"`
}

var People = make(map[uuid.UUID]Person)

func nickHasAlreadyBeenTaken (newPerson *Person) error {
  for _, v := range People {
    if v.Nick == newPerson.Nick {
      return errors.New("Nick has already been taken.")
    }
  }
  return nil
}

func personHasValidStringLength (newPerson *Person, w http.ResponseWriter) error {
  maxNameFieldLength := 100
  maxNickFieldLength := 32
  maxStackFieldLength := 32

  if len(newPerson.Name) > maxNameFieldLength {
    return errors.New("Name field length must be less than 100 characters.")
  }

  if len(newPerson.Nick) > maxNickFieldLength {
    return errors.New("Nick field length must be less than 100 characters.")
  }

  for _, stack := range newPerson.Stack {
    if len(stack) > maxStackFieldLength {
      return errors.New("Stack field length must be less than 100 characters.")
    } 
  }
  return nil
}

func createNewPerson(w http.ResponseWriter, r *http.Request) {
  var newPerson Person
  body, err := io.ReadAll(r.Body)
  if err != nil {
    panic(err)
  }
  err = json.Unmarshal(body, &newPerson)
  if err != nil {
    switch err.(type) {
    case *json.UnmarshalTypeError:
      fmt.Println("Person fields must be of type string.")
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Person field type must be string."))
      return
    default:
      fmt.Println("Cannot create user. DEFAULT error.")
      w.WriteHeader(http.StatusUnprocessableEntity)
      w.Write([]byte("Cannot create user. Please check the provided values."))
      return
    }
  }
  err = personHasValidStringLength(&newPerson, w)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
    return
  }
  err = nickHasAlreadyBeenTaken(&newPerson)
  if err != nil {
    w.WriteHeader(http.StatusUnprocessableEntity)
    w.Write([]byte(err.Error()))
    return
  }
  newPerson.Id, _ = uuid.NewV7()
  People[newPerson.Id] = newPerson
  response := PersonCreatedResponse{
    Message: "Person created successfully!",
  }
  w.Header().Set("Content-Type", "application/json")
  w.Header().Set("Location", fmt.Sprintf("/pessoas/%s", newPerson.Id.String()))
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(response)
}

func getPersonById(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  idToUUID, _ := uuid.Parse(id)
  person, ok := People[idToUUID]
  if ok {
    response := GetPersonResponse{
      PersonReturned: person,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
    return
  }
  w.WriteHeader(http.StatusNotFound)
  w.Write([]byte("Person not found."))
}

func searchPatternInPerson(pattern string, person Person, sliceOfFoundPeople *[]Person) {
  if strings.Contains(person.Name, pattern) {
    *sliceOfFoundPeople = append(*sliceOfFoundPeople, person)
  }

  if strings.Contains(person.Nick, pattern) {
    *sliceOfFoundPeople = append(*sliceOfFoundPeople, person)
  }

  if strings.Contains(person.Birth, pattern) {
    *sliceOfFoundPeople = append(*sliceOfFoundPeople, person)
  }

  for _, lang := range person.Stack {
    if strings.Contains(lang, pattern) {
      *sliceOfFoundPeople = append(*sliceOfFoundPeople, person)
    }
  }
}

func getPersonBySearchPattern(w http.ResponseWriter, r *http.Request) {
  pattern := r.URL.Query().Get("t")
  if len(pattern) == 0 {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Not pattern was providede"))
    return
  }
  foundPeople := []Person{}
  for _, person := range People {
      searchPatternInPerson(pattern, person, &foundPeople)
  }
  if len(foundPeople) <= 0 {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("[]"))
    return
  }
  j, _ := json.Marshal(foundPeople)
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  w.Write(j)
}

func getPeopleCount(w http.ResponseWriter, r *http.Request) {
  response := PeopleCountResponse{
    Count: len(People),
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(response)
}

func main() {
  r := chi.NewRouter()
  r.Use(middleware.Logger)
  r.Post("/pessoas", createNewPerson)
  r.Get("/pessoas", getPersonBySearchPattern)
  r.Get("/pessoas/{id}", getPersonById)
  r.Get("/contagem-pessoas", getPeopleCount)
  http.ListenAndServe(":9999", r)
}
