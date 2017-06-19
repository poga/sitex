package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileServer serves a directory to the web using HTTP.
// It works like http.FileServer, but without directory index (returns 404 when accessing a directory).
type FileServer struct {
	WorkingDir string
}

func (s FileServer) Match(r *http.Request) bool {
	return !strings.HasSuffix(r.URL.Path, "/")
}

// ServeHTTP Serving static files without director index
func (s FileServer) Handle(w http.ResponseWriter, r *http.Request) bool {
	if !s.Match(r) {
		return true
	}

	path := r.URL.Path

	path = filepath.Join(s.WorkingDir, path[1:])
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		http.ServeFile(w, r, path)
		return false
	}
	return true
}
