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
	router MainRouter
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

	var headerRouters []HeaderRouter
	data, err := ioutil.ReadFile(headerConfig)
	if err == nil {
		headerRouters, err = loadHeaderRouters(directory, data)
		if err != nil {
			return nil, err
		}
	}
	data, err = ioutil.ReadFile(redirectConfig)
	shadowingRouter := httprouter.New()
	nonShadowingRouter := httprouter.New()
	if err == nil {
		redirectRoutes, err := loadRedirectRoutes(directory, data)
		if err != nil {
			return nil, err
		}
		for _, route := range redirectRoutes {
			if route.Shadowing {
				route.HookTo(shadowingRouter)
			} else {
				route.HookTo(nonShadowingRouter)
			}
		}
	}
	fileServer := FileServer{directory}
	mainRouter := MainRouter{headerRouters, shadowingRouter, nonShadowingRouter, fileServer}

	return &Server{mainRouter}, nil
}

func loadHeaderRouters(directory string, config []byte) ([]HeaderRouter, error) {
	return NewHeaderRouters(config)
}

func loadRedirectRoutes(directory string, config []byte) ([]*Route, error) {
	routes := make([]*Route, 0)
	lines := bytes.Split(config, []byte("\n"))
	for _, line := range lines {
		route, err := NewRoute(directory, line)
		if err != nil {
			return nil, err
		}
		// comment line
		if route == nil {
			continue
		}
		routes = append(routes, route)
	}
	return routes, nil
}
