package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

// Serve serves a directory with respect to _redirects config
func Serve(workingDir string, addr string) error {
	redirectConfig := filepath.Join(workingDir, "_redirects")
	// if there's no redirect file, just serve static files
	if _, err := os.Stat(redirectConfig); os.IsNotExist(err) {
		http.Handle("/", http.FileServer(http.Dir(".")))
		return http.ListenAndServe(addr, nil)
	}

	data, err := ioutil.ReadFile(redirectConfig)
	if err != nil {
		return err
	}

	router := httprouter.New()
	// define route line by line
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		route, err := NewRoute(workingDir, line)
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
	router.NotFound = FallbackHandler{}

	return http.ListenAndServe(addr, router)
}
