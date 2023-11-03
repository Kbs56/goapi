package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Err string
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Err: err.Error()})
		}
	}
}

func (s *Server) handleGetAllUsers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return writeJSON(
			w,
			http.StatusMethodNotAllowed,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /get", r.Method)},
		)
	}
	users, err := s.pgdb.getAllUsers()
	if err != nil {
		return writeJSON(
			w,
			http.StatusInternalServerError,
			apiError{"We are experinecing difficulties at the moment..."},
		)
	}
	writeJSON(w, http.StatusOK, users)
	return nil
}

func (s *Server) handlgetGetUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("%s method not allowed for endpoint /get", r.Method)},
		)
	}
	// get id from request param
	u, err := s.pgdb.getUser(1)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, u)
}

func (conn *PostgresDB) getAllUsers() ([]*User, error) {
	sqlStatement := `
  select * from customer
  `
	rows, err := conn.db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.FirstName, &user.LastName)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (conn *PostgresDB) getUser(id int) (*User, error) {
	u := User{}
	err := conn.db.QueryRow("select * from customer where id = $1", id).
		Scan(&u.Id, &u.FirstName, &u.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		} else {
			panic(err)
		}
	}
	return &u, nil
}

func (conn *PostgresDB) insertUser(u *User) error {
	sqlStatement := `
  insert into customer (id, first_name, last_name)
  values ($1, $2, $3)
  `
	_, err := conn.db.Exec(sqlStatement, u.Id, u.FirstName, u.LastName)
	if err != nil {
		panic(err)
	}

	fmt.Println(u, "inserted into database")
	return nil
}

func (server *Server) run() {
	fmt.Println("Service started on port", server.listenAddr)
	http.Handle("/get", makeHTTPHandler(server.handleGetAllUsers))
	http.Handle("/getUser", makeHTTPHandler(server.handlgetGetUser))
	http.ListenAndServe(server.listenAddr, nil)
}

func ConnectDB() (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
		os.Getenv("pghost"),
		os.Getenv("pgport"),
		os.Getenv("pguser"),
		os.Getenv("pgpass"),
		os.Getenv("pgdbname"),
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{
		db: db,
	}, nil
}

func main() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatal("Error with connecting to DB")
	}

	server := Server{
		listenAddr: ":3000",
		pgdb:       db,
	}

	server.run()
}
