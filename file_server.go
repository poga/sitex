package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

// FileServer serves a directory to the web using HTTP.
// It works like http.FileServer, but without directory index (returns 404 when accessing a directory).
type FileServer struct {
	WorkingDir    string
	HeaderRouters []HeaderRouter
}

// ServeHTTP Serving static files without director index
func (s FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if len(s.HeaderRouters) > 0 {
		for _, headerRouter := range s.HeaderRouters {
			err := headerRouter.Handle(w, r, nil)
			if err != nil {
				return
			}
		}
	}

	if strings.HasSuffix(path, "/") {
		w.WriteHeader(404)
		return
	}

	path = filepath.Join(s.WorkingDir, path[1:])
	http.ServeFile(w, r, path)
}
