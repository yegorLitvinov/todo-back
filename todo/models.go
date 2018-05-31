package todo

import (
	"fmt"
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

func SetupDB() {
	dbHost := "localhost"
	if gin.Mode() == gin.ReleaseMode {
		dbHost = "postgres"
	}
	var err error
	db, err = gorm.Open(
		"postgres",
		fmt.Sprintf("host=%s port=5432 user=todo dbname=todo password=password sslmode=disable", dbHost),
	)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&User{}, &Tag{}, &Todo{})
}
