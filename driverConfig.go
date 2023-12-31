package main

import (
  "fmt"
)

const PostgresDriver = "postgres"

const Host = "localhost"

const Port = "5432"

const User = "rinha"

const Password = "rinha"

const DbName = "rinha"

const TableName = "people"

var DataSourceName = fmt.Sprintf("host=%s port=%s user=%s "+
    "password=%s dbname=%s sslmode=disable", Host, Port, User, Password, DbName)
