package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"path/filepath"

	"fmt"

	"github.com/julienschmidt/httprouter"
)

// Serve serves a directory with respect to _redirects config
func Serve(workingDir string, addr string) error {
	redirectConfig := filepath.Join(workingDir, "_redirects")
	headerConfig := filepath.Join(workingDir, "_headers")

	router := httprouter.New()

	var headerRouter *HeaderRouter
	data, err := ioutil.ReadFile(headerConfig)
	if err == nil {
		headerRouter, err = loadHeaderConfig(workingDir, data)
		if err != nil {
			return err
		}
	}
	data, err = ioutil.ReadFile(redirectConfig)
	if err == nil {
		loadRedirectConfig(workingDir, router, headerRouter, data)
	}

	router.NotFound = FileServer{workingDir, headerRouter}

	fmt.Printf("Serving %s at %s\n", workingDir, addr)
	return http.ListenAndServe(addr, router)
}

func loadHeaderConfig(workingDir string, config []byte) (*HeaderRouter, error) {
	return NewHeaderRouter(config)
}

func loadRedirectConfig(workingDir string, router *httprouter.Router, headerRouter *HeaderRouter, config []byte) error {
	lines := bytes.Split(config, []byte("\n"))
	for _, line := range lines {
		route, err := NewRoute(workingDir, line, headerRouter)
		if err != nil {
			return err
		}
		// comment line
		if route == nil {
			continue
		}
		if route.IsProxy() {
			// if it's a proxy, we just define the route on all method
			methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
			for _, method := range methods {
				router.Handle(method, route.Match, route.Handler)
			}
		} else {
			router.GET(route.Match, route.Handler)
		}
	}
	return nil
}
