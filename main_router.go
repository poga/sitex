package main

import (
	"net/http"
)

type MainRouter struct {
	headers               []Header
	shadowingRedirects    []*Redirect
	nonShadowingRedirects []*Redirect
	fileServer            FileServer
}

func (main MainRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, header := range main.headers {
		next, err := header.Handle(w, r)
		if err != nil {
			return
		}

		if !next {
			return
		}
	}

	for _, redirect := range main.shadowingRedirects {
		next, err := redirect.Handle(w, r)
		if err != nil {
			return
		}

		if !next {
			return
		}
	}

	next, err := main.fileServer.Handle(w, r)
	if err != nil {
		return
	}
	if !next {
		return
	}

	for _, redirect := range main.nonShadowingRedirects {
		next, err := redirect.Handle(w, r)
		if err != nil {
			return
		}

		if !next {
			return
		}
	}

	w.WriteHeader(404)
}
