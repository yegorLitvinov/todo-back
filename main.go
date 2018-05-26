package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
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
	Todos     []Todo    `gorm:"foreignkey:User" json:"-"`
}

type Todo struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Text      string    `gorm:"not null" json:"text" binding:"required"`
	Order     uint      `gorm:"not null;unique;auto_increment" json:"order"`
	Completed bool      `gorm:"not null" json:"completed"`
	User      uint      `gorm:"not null" json:"-"`
	Tag       uint      `json:"tag"`
}

type Tag struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	Text      string    `gorm:"not null" json:"text"`
	Todos     []Todo    `gorm:"foreignkey:Tag" json:"-"`
}

func listTodos(c *gin.Context) {
	user := getUserFromContext(c)
	fmt.Println(user.ID)
	var todos []Todo
	// todos := make([]Todo, 0, 100)
	err := db.Where(&Todo{User: user.ID}).Find(&todos).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	user := getUserFromContext(c)
	todo.User = user.ID
	if err := db.Create(&todo).Error; err != nil {
		panic(err)
	}
	c.JSON(http.StatusCreated, todo)
}

func updateTodo(c *gin.Context) {
	var todo Todo
	// get todo
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not found"})
		return
	}
	user := getUserFromContext(c)
	err = db.Where(&Todo{User: user.ID, ID: uint(id)}).First(&todo).Error
	if gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not found"})
		return
	}
	if err != nil {
		panic(err)
	}
	// validate
	// TODO: order changes
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := db.Save(&todo).Error; err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, todo)
}

type Credentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func login(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var user User
	if err := db.Where("email = ? and password = ?", creds.Email, creds.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": gin.H{"Tag": "Wrong password / No such user"}})
		return
	}
	setSessionUser(c, user.ID)
	c.JSON(http.StatusOK, user)
}

func logout(c *gin.Context) {
	deleteSessionUser(c)
	c.JSON(http.StatusOK, gin.H{"Info": "You are successfully logged out"})
}

func signup(c *gin.Context) {
	type CredentialsConfirm struct {
		Credentials
		ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Credentials.Password"`
	}
	var creds CredentialsConfirm
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var user User
	user.Email = creds.Email
	user.Password = creds.Password
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": gin.H{"Tag": "User already exists"}})
		return
	}
	setSessionUser(c, user.ID)
	c.JSON(http.StatusCreated, gin.H{"Info": "The user successfully created"})
}

func getSessionUser(c *gin.Context) (User, error) {
	var user User
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(uint)
	if !ok {
		return user, errors.New("Session is empty")
	}
	if err := db.First(&user, userID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return user, errors.New("User not found")
		}
		panic(err)
	}
	return user, nil
}

func setSessionUser(c *gin.Context, userID uint) {
	session := sessions.Default(c)
	session.Set("userID", userID)
	if err := session.Save(); err != nil {
		panic(err)
	}
}

func deleteSessionUser(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userID")
}

func getUserFromContext(c *gin.Context) User {
	value, exist := c.Get("user")
	if !exist {
		panic(errors.New("session user not exists"))
	}
	user, ok := value.(User)
	if !ok {
		panic(errors.New("user in the context is not really user"))
	}
	return user
}

func authRequiredMiddleware(c *gin.Context) {
	user, err := getSessionUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
		c.Abort()
	}
	c.Set("user", user)
	c.Next()
}

func main() {
	var err error
	// DB
	db, err = gorm.Open(
		"postgres",
		"host=localhost port=5432 user=todo dbname=todo password=password sslmode=disable",
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.AutoMigrate(&User{}, &Tag{}, &Todo{})

	// Session
	var store redis.Store
	store, err = redis.NewStore(10, "tcp", "localhost:6379", "", []byte("LFoNBWW394@#$#d"))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{MaxAge: 60 * 60 * 1, Path: "/"})

	// Routes/Middleware
	router := gin.Default()
	sessionMiddleware := sessions.Sessions("x-session", store)
	router.Use(sessionMiddleware)

	apiGroup := router.Group("/api/v1/")
	todoGroup := apiGroup.Group("/todos/", authRequiredMiddleware)
	authGroup := apiGroup.Group("/auth/")

	authGroup.POST("/login/", login)
	authGroup.POST("/signup/", signup)
	authGroup.POST("/logout/", authRequiredMiddleware, logout)

	todoGroup.GET("/", listTodos)
	todoGroup.POST("/", createTodo)
	todoGroup.PUT("/:id/", updateTodo)

	router.Run(":4000")
}
