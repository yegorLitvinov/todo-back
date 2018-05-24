package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Todos     []Todo    `gorm:"foreignkey:User" json:"todos"`
}

type Todo struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null"json:"updatedAt"`
	Text      string    `gorm:"not null" json:"text" binding:"required"`
	Order     uint      `gorm:"not null;unique;auto_increment" json:"order"`
	Completed bool      `gorm:"not null" json:"completed"`
	User      *uint     `json:"user"`
	Tag       *uint     `json:"tag"`
}

type Tag struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	Text      string    `gorm:"not null" json:"text"`
	Todos     []Todo    `gorm:"foreignkey:Tag" json:"-"`
}

func listTodos(c *gin.Context) {
	var todos []Todo
	db.Find(&todos)
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	db.Create(&todo)
	c.JSON(http.StatusCreated, todo)
}

func updateTodo(c *gin.Context) {
	var todo Todo
	id := c.Param("id")
	if err := db.First(&todo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, "")
		return
	}
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	db.Save(&todo)
	c.JSON(http.StatusOK, todo)
}

func main() {
	var err error
	db, err = gorm.Open(
		"postgres",
		"host=localhost port=5432 user=todo dbname=todo password=password sslmode=disable",
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.AutoMigrate(&User{}, &Tag{}, &Todo{})
	router := gin.Default()
	apiGroup := router.Group("/api/v1/")
	todoGroup := apiGroup.Group("/todos/")

	todoGroup.GET("/", listTodos)
	todoGroup.POST("/", createTodo)
	todoGroup.PUT("/:id/", updateTodo)

	router.Run(":4000")
}
