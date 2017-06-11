package main

import (
	"io/ioutil"

	"bytes"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	*httprouter.Router
}

func NewRouter(path string) (*Router, error) {
	r := httprouter.New()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		ParseRedirectRule(line)
	}

	return &Router{r}, nil
}
