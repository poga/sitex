package main

import "net/http"

type middleware interface {
	// Match returns true if the request is a match to the middleware
	Match(*http.Request) bool
	// Handle process the request, send response,
	// then returns whether the middleware chain should go on, and error
	Handle(w http.ResponseWriter, r *http.Request) (bool, error)
}
