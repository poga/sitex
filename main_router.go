package main

import (
	"net/http"
)

type MainRouter struct {
	headers               []middleware
	shadowingRedirects    []middleware
	nonShadowingRedirects []middleware
	fileServer            middleware
}

func (main MainRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	run([][]middleware{
		main.headers,
		main.shadowingRedirects,
		{main.fileServer},
		main.nonShadowingRedirects,
	}, w, r)
}

func run(layers [][]middleware, w http.ResponseWriter, r *http.Request) {
	for _, layer := range layers {
		for _, mw := range layer {
			next := mw.Handle(w, r)
			if !next {
				return
			}
		}
	}

	w.WriteHeader(404)
}
