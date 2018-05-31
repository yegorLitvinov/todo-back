package main

import (
	"github.com/yegorLitvinov/todo-back/todo"
)

func main() {
	todo.SetupDB()
	store := todo.SetupSessionStore()
	router := todo.SetupAPIRouter(store)
	router.Run(":4000")
}
