package todo

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Need it to use postgres
)

// User which uses the app
type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Todos     []Todo    `gorm:"foreignkey:User" json:"-"`
}

func (u *User) Credentials() *Credentials {
	return &Credentials{Email: u.Email, Password: u.Password}
}

// Todo item
type Todo struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Text      string    `gorm:"not null" json:"text" binding:"required"`
	Order     int       `gorm:"not null;unique_index:idx_order_user" json:"order"`
	Completed bool      `gorm:"not null" json:"completed"`
	User      uint      `gorm:"not null;unique_index:idx_order_user" json:"-"`
	Tag       uint      `json:"tag"`
}

// Tag for todo describes the category
type Tag struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	Text      string    `gorm:"not null" json:"text"`
	Todos     []Todo    `gorm:"foreignkey:Tag" json:"-"`
}

// SetupDB inits the database connection
func SetupDB() *gorm.DB {
	dbHost := "localhost"
	logMode := true
	if gin.Mode() == gin.ReleaseMode {
		dbHost = "postgres"
		logMode = false
	}
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf("host=%s port=5432 user=todo dbname=todo password=password sslmode=disable", dbHost),
	)
	if err != nil {
		panic(err)
	}
	db.LogMode(logMode)
	db.AutoMigrate(&User{}, &Tag{}, &Todo{})
	return db
}

// DBMiddleware saves db in context
func DBMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// GetDB get db fom context with errors handling
func GetDBFromContext(c *gin.Context) *gorm.DB {
	contextInstance, exists := c.Get("db")
	if !exists {
		panic(errors.New("No db in context"))
	}
	var ok bool
	var db *gorm.DB
	if db, ok = contextInstance.(*gorm.DB); !ok {
		panic("Can't cast to *gorm.DB")
	}
	return db
}

func selectNextTodoOrder(u *User, db *gorm.DB) int {
	var order int
	row := db.Table("todos").Where(&Todo{User: u.ID}).Select("max(\"order\")").Row()
	if err := row.Scan(&order); err != nil {
		return 0
	}
	return order + 1
}

type byOrder []Todo

func (a byOrder) Len() int               { return len(a) }
func (a byOrder) Swap(i int, j int)      { a[i], a[j] = a[j], a[i] }
func (a byOrder) Less(i int, j int) bool { return a[i].Order < a[j].Order }

func (todo *Todo) reorder(todos []Todo, oldOrder int, tx *gorm.DB) error {
	newOrder := todo.Order
	minOrder := oldOrder
	maxOrder := newOrder
	sign := -1
	sort.Sort(byOrder(todos))
	if newOrder < oldOrder {
		minOrder = newOrder
		maxOrder = oldOrder
		sign = 1
		sort.Sort(sort.Reverse(byOrder(todos)))
	}
	todo.Order = 32767
	if err := tx.Save(todo).Error; err != nil {
		return err
	}
	for _, t := range todos {
		if t.Order != oldOrder && t.Order >= minOrder && t.Order <= maxOrder {
			t.Order += sign
			if err := tx.Save(&t).Error; err != nil {
				return err
			}
		}
	}
	todo.Order = newOrder
	if err := tx.Save(todo).Error; err != nil {
		return err
	}
	return nil
}
