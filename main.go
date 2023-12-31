package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

var database Repository
func init() {
  repository := &Repository{}
  var err error
  database.db, err = repository.connectToDB()
  if err != nil {
    panic(err.Error())
  }
}

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
	Nick  string   `json:"apelido"`
	Name  string   `json:"nome"`
	Birth string   `json:"nascimento"`
	Stack []string `json:"stack"`
}

func personHasValidStringLength(newPerson *Person, w http.ResponseWriter) error {
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
	newPerson.Id, _ = uuid.NewV7()
  err = database.createNewPerson(newPerson)
  if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
  }
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
  person, err := database.findPersonById(idToUUID)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(err.Error()))
    return
  }
  response := GetPersonResponse{
    PersonReturned: person[0],
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(response)
}

func getPersonBySearchPattern(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("t")
	if len(pattern) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not pattern was provided"))
		return
	}
  person, err := database.findPersonByPattern(pattern)
  if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
  }
	if len(person) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}
	j, _ := json.Marshal(person)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func getPeopleCount(w http.ResponseWriter, r *http.Request) {
  peopleCount, err := database.getPeopleCount()
  if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("0"))
  }
	response := PeopleCountResponse{
		Count: peopleCount,
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
