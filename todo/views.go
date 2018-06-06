package todo

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func listTodos(c *gin.Context) {
	user := getUserFromContext(c)
	var todos []Todo
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
	todo.Order = user.selectNextTodoOrder()
	if err := db.Create(&todo).Error; err != nil {
		panic(err)
	}
	c.JSON(http.StatusCreated, todo)
}

func updateTodo(c *gin.Context) {
	tx := db.Begin()
	var todo Todo
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not found"})
		return
	}
	user := getUserFromContext(c)
	err = tx.Where(&Todo{User: user.ID, ID: uint(id)}).First(&todo).Error
	if gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not found"})
		return
	}
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	oldOrder := todo.Order
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if oldOrder != todo.Order {
		// reordering
		var todos []Todo
		if err := db.Where(&Todo{User: user.ID}).Find(&todos).Error; err != nil {
			tx.Rollback()
			panic(err)
		}
		if err := todo.reorder(todos, oldOrder, tx); err != nil {
			tx.Rollback()
			panic(err)
		}
	}
	if err := tx.Save(&todo).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	tx.Commit()
	c.JSON(http.StatusOK, todo)
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

// SetupAPIRouter init routes and middlewares
func SetupAPIRouter(store sessions.Store) *gin.Engine {
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
	return router
}
