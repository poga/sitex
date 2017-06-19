package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileServer(t *testing.T) {
	server := FileServer{"./example"}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.Handle(rec, req)
	require.Equal(t, 200, rec.Code)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test.json", nil)
	server.Handle(rec, req)
	require.Equal(t, 200, rec.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", rec.Body.String())

	// won't return 404 if file doesn't exist
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	server.Handle(rec, req)
	require.Equal(t, 200, rec.Code)
}
