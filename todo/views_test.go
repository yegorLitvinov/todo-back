package todo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func beforeEach() {
	if db == nil {
		SetupDB()
	}
	db.Delete(Todo{})
	db.Delete(User{})
	db.Delete(Tag{})
}

func performRequest(method, path string, data *interface{}) *httptest.ResponseRecorder {
	store := SetupSessionStore()
	r := SetupAPIRouter(store)
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
	r.ServeHTTP(w, req)
	return w
}

func TestListTodo(t *testing.T) {
	beforeEach()

	todo := Todo{Text: "Hello"}
	db.Create(&todo)

	w := performRequest("GET", "/api/v1/todos/", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
