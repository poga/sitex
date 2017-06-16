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
	router   *httprouter.Router
	listener net.Listener
}

// Start starts the server
func (s *Server) Start() {
	http.Serve(s.listener, s.router)
}

// NewServer looks for `_headers` and `_redirects` files in the workingDir, parse rules, and return a Server
func NewServer(workingDir string, addr string) (*Server, error) {
	redirectConfig := filepath.Join(workingDir, "_redirects")
	headerConfig := filepath.Join(workingDir, "_headers")

	router := httprouter.New()

	var headerRouters []HeaderRouter
	data, err := ioutil.ReadFile(headerConfig)
	if err == nil {
		headerRouters, err = loadHeaderConfig(workingDir, data)
		if err != nil {
			return nil, err
		}
	}
	data, err = ioutil.ReadFile(redirectConfig)
	if err == nil {
		loadRedirectConfig(workingDir, router, headerRouters, data)
	}

	router.NotFound = FileServer{workingDir, headerRouters}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{router, listener}, nil
}

func loadHeaderConfig(workingDir string, config []byte) ([]HeaderRouter, error) {
	return NewHeaderRouters(config)
}

func loadRedirectConfig(workingDir string, router *httprouter.Router, headerRouters []HeaderRouter, config []byte) error {
	lines := bytes.Split(config, []byte("\n"))
	for _, line := range lines {
		route, err := NewRoute(workingDir, line, headerRouters)
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
