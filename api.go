package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

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

func (s *Server) handleUpdateEmail(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /updateEmail", r.Method)},
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
	u, err = s.pgdb.updateEmail(u)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, &u)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return writeJSON(
			w,
			http.StatusBadRequest,
			apiError{Err: fmt.Sprintf("Method %s not allowed for endpoint /delete", r.Method)},
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
	err = s.pgdb.deleteUser(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeJSON(
				w,
				http.StatusBadRequest,
				apiError{Err: fmt.Sprintf("No user with id %d", id)},
			)
		}
		return writeJSON(
			w,
			http.StatusInternalServerError,
			apiError{Err: "Error deleting user"},
		)
	}
	return writeJSON(w, http.StatusNoContent, nil)
}
