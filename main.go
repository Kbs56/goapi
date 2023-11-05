package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func (server *Server) run() {
	fmt.Println("Service started on port", server.listenAddr)
	http.Handle("/create", makeHTTPHandler(server.handleCreateUser))
	http.Handle("/getUsers", makeHTTPHandler(server.handleGetAllUsers))
	http.Handle("/getUser", makeHTTPHandler(server.handleGetUser))
	http.Handle("/updateEmail", makeHTTPHandler(server.handleUpdateEmail))
	http.Handle("/delete", makeHTTPHandler(server.handleDeleteUser))
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
