package todo

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func main() {
	var err error
	// DB
	dbHost := "localhost"
	if gin.Mode() == gin.ReleaseMode {
		dbHost = "postgres"
	}
	db, err = gorm.Open(
		"postgres",
		fmt.Sprintf("host=%s port=5432 user=todo dbname=todo password=password sslmode=disable", dbHost),
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.AutoMigrate(&User{}, &Tag{}, &Todo{})

	// Session
	var store redis.Store
	redisHost := "localhost"
	if gin.Mode() == gin.ReleaseMode {
		redisHost = "redis"
	}
	store, err = redis.NewStore(10, "tcp", fmt.Sprintf("%s:6379", redisHost), "", []byte("LFoNBWW394@#$#d"))
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
