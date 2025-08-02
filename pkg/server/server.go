package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"go-todo-list/pkg/api"
)

func StartServer(webDir string) {

	//if TODO_PORT variable of empty use port 7540
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// FileServer for files from /web folder
	server := http.FileServer(http.Dir(webDir))
	http.Handle("/", server)

	serverAddress := fmt.Sprintf(":%s", port)
	log.Printf("Directory with files is /%s. Starting server on port %s. Link is http://localhost%s/\n", webDir, port, serverAddress)

	//add handlers
	api.Init()

	//catch any errors
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
