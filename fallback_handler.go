package main

import "net/http"

import "strings"

type FallbackHandler struct {
}

// Serving static files without director index
func (h FallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/") {
		w.WriteHeader(404)
		return
	}
	http.ServeFile(w, r, path[1:])
}
