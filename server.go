package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func Serve(path string, addr string) error {
	redirectConfig := filepath.Join(path, "_redirects")
	// if there's no redirect file, just serve static files
	if _, err := os.Stat(redirectConfig); os.IsNotExist(err) {
		http.Handle("/", http.FileServer(http.Dir(".")))
		return http.ListenAndServe(addr, nil)
	}

	data, err := ioutil.ReadFile(redirectConfig)
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	// define route line by line
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		route, err := ParseRedirectRule(path, line)
		if err != nil {
			return err
		}
		// comment line
		if route == nil {
			continue
		}
		router.GET(route.Match, route.Handler)
	}
	router.NotFound = FallbackHandler{}

	return http.ListenAndServe(addr, router)
}
