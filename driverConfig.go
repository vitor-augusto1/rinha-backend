package main

import (
  "fmt"
)

const PostgresDriver = "postgres"

const Host = "172.18.0.2"

const Port = "5432"

const User = "rinha"

const Password = "rinha"

const DbName = "rinha"

const TableName = "people"

var DataSourceName = fmt.Sprintf("host=%s port=%s user=%s "+
    "password=%s dbname=%s sslmode=disable", Host, Port, User, Password, DbName)
