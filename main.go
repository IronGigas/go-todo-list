package main

import (
	"go-todo-list/pkg/db"
	"go-todo-list/pkg/server"
	"os"
)

func main() {

	//create db or initialize
	db.Init(os.Getenv("TODO_DBFILE"))
	//close db after work is done
	defer db.Db.Close()
	//start server
	server.StartServer("web")

}
