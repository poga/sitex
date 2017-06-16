package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

// FileServer works similar to http.FileServer, but without directory index
type FileServer struct {
	WorkingDir    string
	HeaderRouters []HeaderRouter
}

// ServeHTTP Serving static files without director index
func (s FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if len(s.HeaderRouters) > 0 {
		for _, headerRouter := range s.HeaderRouters {
			handle, _, _ := headerRouter.Lookup("GET", path)
			if handle != nil {
				handle(w, r, nil)
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
