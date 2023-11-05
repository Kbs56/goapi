package main

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

type Server struct {
	listenAddr string
	pgdb       *PostgresDB
}

type PostgresDB struct {
	db *sql.DB
}

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Err string
}
