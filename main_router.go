package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type MainRouter struct {
	headerRouters      []HeaderRouter
	shadowingRouter    *httprouter.Router
	nonShadowingRouter *httprouter.Router
	fileServer         FileServer
}

func (main MainRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, headerRouter := range main.headerRouters {
		handler, _, _ := headerRouter.Lookup(r.Method, r.URL.Path)
		if handler != nil {
			err := headerRouter.Handle(w, r, nil)
			if err != nil {
				return
			}
		}
	}

	handler, _, _ := main.shadowingRouter.Lookup(r.Method, r.URL.Path)
	if handler != nil {
		handler(w, r, nil)
		return
	}

	err := main.fileServer.ServeHTTP(w, r)
	if err != nil {
		return
	}

	handler, _, _ = main.nonShadowingRouter.Lookup(r.Method, r.URL.Path)
	if handler != nil {
		handler(w, r, nil)
		return
	}

	w.WriteHeader(404)
}
