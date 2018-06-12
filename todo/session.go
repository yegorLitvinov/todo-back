package todo

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getSessionUser(c *gin.Context) (User, error) {
	var user User
	session := sessions.Default(c)
	userID, ok := session.Get("userID").(uint)
	if !ok {
		return user, errors.New("Session is empty")
	}
	db := GetDBFromContext(c)
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

// SetupSessionStore init the session
func SetupSessionStore() sessions.Store {
	var store redis.Store
	redisHost := "localhost"
	if gin.Mode() == gin.ReleaseMode {
		redisHost = "redis"
	}
	store, err := redis.NewStore(10, "tcp", fmt.Sprintf("%s:6379", redisHost), "", []byte("LFoNBWW394@#$#d"))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{MaxAge: 60 * 60 * 1, Path: "/"})
	return store
}
