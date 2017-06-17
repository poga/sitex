package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"path/filepath"

	"net"

	"github.com/julienschmidt/httprouter"
)

// Server is an instance of SiteX server
type Server struct {
	router *httprouter.Router
}

// Start starts the server
func (s *Server) Start(listener net.Listener) {
	http.Serve(listener, s.router)
}

// NewServer creates a new server serving given directory.
// It follows the rules defined in `_redirects` and `_headers` files.
func NewServer(directory string) (*Server, error) {
	redirectConfig := filepath.Join(directory, "_redirects")
	headerConfig := filepath.Join(directory, "_headers")

	router := httprouter.New()

	var headerRouters []HeaderRouter
	data, err := ioutil.ReadFile(headerConfig)
	if err == nil {
		headerRouters, err = loadHeaderConfig(directory, data)
		if err != nil {
			return nil, err
		}
	}
	data, err = ioutil.ReadFile(redirectConfig)
	if err == nil {
		loadRedirectConfig(directory, router, headerRouters, data)
	}

	router.NotFound = FileServer{directory, headerRouters}

	return &Server{router}, nil
}

func loadHeaderConfig(directory string, config []byte) ([]HeaderRouter, error) {
	return NewHeaderRouters(config)
}

func loadRedirectConfig(directory string, router *httprouter.Router, headerRouters []HeaderRouter, config []byte) error {
	lines := bytes.Split(config, []byte("\n"))
	for _, line := range lines {
		route, err := NewRoute(directory, line, headerRouters)
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
