package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileServer(t *testing.T) {
	handler := FileServer{"./example", nil}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 404, rec.Code)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test.json", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", rec.Body.String())

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 404, rec.Code)
}

func TestFileServerWithHeaderRouter(t *testing.T) {
	config := `
/test.json
	X-TEST: hello
	`
	router, _ := NewHeaderRouters([]byte(config))

	handler := FileServer{"./example", router}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 404, rec.Code)
	require.Equal(t, "", rec.Header().Get("X-TEST"))

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test.json", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", rec.Body.String())
	require.Equal(t, "hello", rec.Header().Get("X-TEST"))

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, 404, rec.Code)
	require.Equal(t, "", rec.Header().Get("X-TEST"))
}
