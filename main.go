package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {
	// if there's no redirect file, just serve static files
	if _, err := os.Stat("_redirects"); os.IsNotExist(err) {
		http.Handle("/", http.FileServer(http.Dir(".")))
		log.Fatal(http.ListenAndServe(":8080", nil))
		return
	}

	data, err := ioutil.ReadFile("_redirects")
	if err != nil {
		log.Fatal(err)
	}
	lines := bytes.Split(data, []byte("\n"))
	router := httprouter.New()
	for _, line := range lines {
		route, err := ParseRedirectRule(line)
		if err != nil {
			log.Fatal(err)
		}
		// comment line
		if route == nil {
			continue
		}
		router.GET(route.Match, route.Handler)
	}
	router.NotFound = FallbackHandler{}

	log.Fatal(http.ListenAndServe(":8080", router))
}
