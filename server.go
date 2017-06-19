package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"path/filepath"

	"net"
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
	var err error

	redirectConfig := filepath.Join(directory, "_redirects")
	headerConfig := filepath.Join(directory, "_headers")

	var headers []middleware
	data, err := ioutil.ReadFile(headerConfig)
	if err == nil {
		headers, err = loadHeaders(directory, data)
		if err != nil {
			return nil, err
		}
	}

	var shadowingRedirects []middleware
	var nonShadowingRedirects []middleware
	data, err = ioutil.ReadFile(redirectConfig)
	if err == nil {
		shadowingRedirects, nonShadowingRedirects, err = loadRedirects(directory, data)
		if err != nil {
			return nil, err
		}
	}
	fileServer := FileServer{directory}
	mainRouter := MainRouter{headers, shadowingRedirects, nonShadowingRedirects, fileServer}

	return &Server{mainRouter}, nil
}

func loadHeaders(directory string, config []byte) ([]middleware, error) {
	return NewHeaders(config)
}

func loadRedirects(directory string, config []byte) ([]middleware, []middleware, error) {
	shadowingRedirects := make([]middleware, 0)
	nonShadowingRedirects := make([]middleware, 0)
	lines := bytes.Split(config, []byte("\n"))
	for _, line := range lines {
		redirect, err := NewRedirect(directory, line)
		if err != nil {
			return nil, nil, err
		}
		// comment line
		if redirect == nil {
			continue
		}
		if redirect.Shadowing {
			shadowingRedirects = append(shadowingRedirects, redirect)
		} else {
			nonShadowingRedirects = append(nonShadowingRedirects, redirect)
		}
	}
	return shadowingRedirects, nonShadowingRedirects, nil
}
