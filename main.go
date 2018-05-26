package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
	User      uint      `gorm:"not null" json:"user"`
	Tag       uint      `json:"tag"`
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Wrong password / No such user"})
		return
	}
	setSessionUser(c, user.ID)
	c.JSON(http.StatusOK, user)
}

func logout(c *gin.Context) {
	value, exist := c.Get("user")
	if !exist {
		panic(errors.New("session user not exists"))
	}
	if _, ok := value.(User); !ok {
		panic(errors.New("user in the session is not really user"))
	}
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": "User already exists"})
		return
	}
	setSessionUser(c, user.ID)
	c.JSON(http.StatusCreated, gin.H{"Info": "The user successfully created"})
}

func getSessionUser(c *gin.Context) *User {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		return nil
	}
	user := new(User)
	if err := db.First(user, userID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		panic(err)
	}
	return user
}

func setSessionUser(c *gin.Context, userID uint) {
	session := sessions.Default(c)
	session.Set("userID", userID)
}

func deleteSessionUser(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userID")
}

func authRequiredMiddleware(c *gin.Context) {
	user := getSessionUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, "")
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
	// store, err = redis.NewStore(10, "tcp", "localhost:6379", "", []byte("LFoNBWW394@#$#d"))
	store = cookie.NewStore([]byte("LFoNBWW394@#$#d"))
	if err != nil {
		panic(err)
	}

	// Routes/Middleware
	router := gin.Default()
	sessionMiddleware := sessions.Sessions("x-session", store)
	// router.Use()

	apiGroup := router.Group("/api/v1/")
	todoGroup := apiGroup.Group("/todos/", authRequiredMiddleware)
	authGroup := apiGroup.Group("/auth/")
	authGroup.Use(sessionMiddleware)

	authGroup.POST("/login/", login)
	authGroup.POST("/signup/", signup)
	authGroup.POST("/logout/", authRequiredMiddleware, logout)

	todoGroup.GET("/", listTodos)
	todoGroup.POST("/", createTodo)
	todoGroup.PUT("/:id/", updateTodo)

	router.Run(":4000")
}
