package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileServer(t *testing.T) {
	handler := FileServer{"./example", nil}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test.json", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", rec.Body.String())

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)
}

func TestFileServerWithHeaderRouter(t *testing.T) {
	config := `
/test.json
	X-TEST: hello
	`
	router, _ := NewHeaderRouter([]byte(config))

	handler := FileServer{"./example", router}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)
	assert.Equal(t, "", rec.Header().Get("X-TEST"))

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test.json", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", rec.Body.String())
	assert.Equal(t, "hello", rec.Header().Get("X-TEST"))

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)
	assert.Equal(t, "", rec.Header().Get("X-TEST"))
}
