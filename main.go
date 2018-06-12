package main

import (
	"github.com/yegorLitvinov/todo-back/todo"
)

func main() {
	db := todo.SetupDB()
	store := todo.SetupSessionStore()
	router := todo.SetupAPIRouter(store, db)
	router.Run(":4000")
}
