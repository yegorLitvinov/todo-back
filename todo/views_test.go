package todo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var testDB *gorm.DB
var testStore sessions.Store
var testRouter *gin.Engine

func beforeEach() {
	if testDB == nil {
		testDB = SetupDB()
	}
	if testStore == nil {
		testStore = SetupSessionStore()
	}
	if testRouter == nil {
		testRouter = SetupAPIRouter(testStore, testDB)
	}
	testDB.Delete(Todo{})
	testDB.Delete(User{})
	testDB.Delete(Tag{})
}

func performRequest(method, path string, data interface{}) *httptest.ResponseRecorder {
	var reader io.Reader
	if data == nil {
		reader = bytes.NewBuffer([]byte{})
	} else {
		body, _ := json.Marshal(data)
		reader = bytes.NewBuffer([]byte(body))
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestLogin(t *testing.T) {
	beforeEach()
	var w *httptest.ResponseRecorder

	user := User{Email: "a@a.com", Password: "1234"}
	testDB.Create(&user)

	w = performRequest("POST", "/api/v1/auth/login/", user.Credentials())
	assert.Equal(t, http.StatusOK, w.Code)

	w = performRequest("POST", "/api/v1/auth/login/", gin.H{"email": "a@a.com", "password": "1234"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = performRequest("POST", "/api/v1/auth/login/", gin.H{"email": "a@a.com", "password": "12345"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = performRequest("POST", "/api/v1/auth/login/", gin.H{"email": "b@a.com", "password": "1234"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnautorized(t *testing.T) {
	beforeEach()

	w := performRequest("GET", "/api/v1/todos/", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
