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
	err := run([][]middleware{
		main.headers,
		main.shadowingRedirects,
		[]middleware{main.fileServer},
		main.nonShadowingRedirects,
	}, w, r)

	if err != nil {
		w.WriteHeader(500)
	}
}

func run(layers [][]middleware, w http.ResponseWriter, r *http.Request) error {
	for _, layer := range layers {
		for _, mw := range layer {
			next, err := mw.Handle(w, r)
			if err != nil {
				return err
			}
			if !next {
				return nil
			}
		}
	}

	w.WriteHeader(404)
	return nil
}
