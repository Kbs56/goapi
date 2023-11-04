package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /getUsers", r.Method)},
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

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /create", r.Method)},
		)
	}
	body := json.NewDecoder(r.Body)
	u := &User{}
	err := body.Decode(u)
	if err != nil {
		return writeJSON(
			w,
			http.StatusInternalServerError,
			apiError{Err: "Error occured creating user"},
		)
	}
	u, err = s.pgdb.createUser(u)
	if err != nil {
		return writeJSON(
			w,
			http.StatusInternalServerError,
			apiError{Err: "Error occured creating user"},
		)
	}
	return writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /getUser", r.Method)},
		)
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: "Please ensure you are passing in a valid ID"},
		)
	}
	u, err := s.pgdb.getUser(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeJSON(w, http.StatusOK, apiError{Err: fmt.Sprintf("No user with ID %d", id)})
		} else {
			return err
		}
	}
	return writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /update", r.Method)},
		)
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: "Please ensure you are passing in a valid ID"},
		)
	}
	u := &User{Id: id}
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return writeJSON(
			w,
			http.StatusInternalServerError,
			apiError{Err: "We are experinecing difficulties"},
		)
	}
	u, err = s.pgdb.updateUser(u)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, &u)
}

func (conn *PostgresDB) updateUser(u *User) (*User, error) {
	sqlStatement := configureQuery(u)
	if err := conn.db.QueryRow(sqlStatement).Scan(&u.Id, &u.FirstName, &u.LastName); err != nil {
		return nil, err
	}
	return u, nil
}

func configureQuery(u *User) string {
	if u.LastName == "" && u.FirstName != "" {
		sqlStatement := fmt.Sprintf(
			"update customer set first_name = '%s' where id = %d returning id, first_name, last_name",
			u.FirstName,
			u.Id,
		)
		return sqlStatement
	} else if u.FirstName == "" && u.LastName != "" {
		sqlStatement := fmt.Sprintf(
			"update customer set last_name = '%s' where id = %d returning id, first_name, last_name",
			u.LastName,
			u.Id,
		)
		return sqlStatement
	} else {
		sqlStatement := fmt.Sprintf(
			"update customer set first_name = '%s', last_name = '%s' where id = %d returning id, first_name, last_name",
			u.FirstName,
			u.LastName,
			u.Id,
		)
		return sqlStatement
	}
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

func (conn *PostgresDB) createUser(u *User) (*User, error) {
	sqlStatement := fmt.Sprintf(
		"INSERT INTO CUSTOMER (FIRST_NAME, LAST_NAME) VALUES ('%s', '%s') RETURNING ID",
		u.FirstName,
		u.LastName,
	)
	var id int
	if err := conn.db.QueryRow(sqlStatement).Scan(&id); err != nil {
		return nil, err
	}
	u.Id = id
	return u, nil
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

func (server *Server) run() {
	fmt.Println("Service started on port", server.listenAddr)
	http.Handle("/create", makeHTTPHandler(server.handleCreateUser))
	http.Handle("/getUsers", makeHTTPHandler(server.handleGetAllUsers))
	http.Handle("/getUser", makeHTTPHandler(server.handleGetUser))
	http.Handle("/update", makeHTTPHandler(server.handleUpdateUser))
	http.ListenAndServe(server.listenAddr, nil)
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
