package main

import (
	"fmt"
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

// ServeHTTP Serving static files without director index
func (s FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path

	if strings.HasSuffix(path, "/") {
		// ignore directory
		return nil
	}

	path = filepath.Join(s.WorkingDir, path[1:])
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		http.ServeFile(w, r, path)
		return fmt.Errorf("File Served")
	}
	return nil
}
