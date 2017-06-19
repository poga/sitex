package main

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"strings"

	"path/filepath"

	"io/ioutil"

	"github.com/julienschmidt/httprouter"
)

// Redirect correspond to a line in the _redirect config
type Redirect struct {
	From       string
	Queries    map[string]string
	StatusCode int
	To         string
	wd         string
	Shadowing  bool
	router     *httprouter.Router
}

func (redirect *Redirect) Match(r *http.Request) bool {
	handle, _, _ := redirect.router.Lookup(r.Method, r.URL.Path)
	if handle == nil {
		return false
	}

	// check if query params matched the request
	if len(redirect.Queries) > 0 {
		queryMatched := true
		for query := range redirect.Queries {
			if r.URL.Query().Get(query) == "" {
				queryMatched = false
				break
			}
		}
		if !queryMatched {
			return false
		}
	}

	return true
}

// Handle handle the request and stop middleware chain if necessary
func (redirect *Redirect) Handle(w http.ResponseWriter, r *http.Request) bool {
	if !redirect.Match(r) {
		return true
	}
	handle, params, _ := redirect.router.Lookup(r.Method, r.URL.Path)
	handle(w, r, params)
	return false
}

// IsProxy returns true if the route is a proxy route.
// A proxy route has a complete URL in its "to" part.
// The route should act as a reverse proxy if it's a proxy route.
func (redirect *Redirect) IsProxy() bool {
	return strings.HasPrefix(redirect.To, "http")
}

// compileRedirectTo returns a string representing the destination of a request
// based on matched route, placeholder, splat, and query params.
func (redirect *Redirect) compileRedirectTo(r *http.Request, ps httprouter.Params) string {
	var pattern = redirect.To
	// is there any splat in the pattern?
	if strings.HasSuffix(pattern, ":splat") {
		splat := ps.ByName("splat")
		splat = splat[1:] // remove "/" at the beginning
		return strings.Replace(pattern, ":splat", splat, 1)
	}

	// does this require query param matching?
	if len(redirect.Queries) > 0 {
		result := pattern
		for query, placeholder := range redirect.Queries {
			result = strings.Replace(result, fmt.Sprintf(":%s", placeholder), r.URL.Query().Get(query), 1)
		}
		return result
	}

	// is there any placeholder in the pattern?
	if strings.Contains(pattern, ":") {
		varName := regexp.MustCompile(":[^/]+")
		vars := varName.FindAllString(pattern, -1)
		result := pattern
		for _, v := range vars {
			name := v[1:]
			result = strings.Replace(result, v, ps.ByName(name), 1)
		}
		return result
	}

	// no replace needed
	return pattern
}

func (redirect *Redirect) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if redirect.StatusCode >= 300 && redirect.StatusCode < 400 {
		http.Redirect(w, r, redirect.compileRedirectTo(r, ps), redirect.StatusCode)
		return
	}

	if redirect.IsProxy() {
		req, err := http.NewRequest(r.Method, redirect.To, r.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		for key, vals := range resp.Header {
			w.Header().Set(key, vals[0])
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(body)
		return
	}

	http.ServeFile(w, r, filepath.Join(redirect.wd, redirect.compileRedirectTo(r, ps)))
}

// NewRedirect returns a route based on given redirect rule.
func NewRedirect(wd string, line []byte) (*Redirect, error) {
	// remove all comments
	comment := regexp.MustCompile("#.+")
	rule := comment.ReplaceAll(line, []byte(""))
	rule = bytes.Trim(rule, " \t")

	// skip empty line or comment line
	if bytes.Compare(rule, []byte{}) == 0 {
		return nil, nil
	}

	// check redirect status code
	fields := bytes.FieldsFunc(rule, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

	// at least two field
	if len(fields) < 2 {
		return nil, fmt.Errorf("Invalid Redirect Rule: %s", line)
	}

	redirect := Redirect{Queries: make(map[string]string), wd: wd, router: httprouter.New()}

	// parse match
	matcher, fields := takeField(fields)
	// if it's a splat route, add a variable name for httprouter
	if strings.HasSuffix(matcher, "*") {
		matcher = matcher + "splat"
	}
	redirect.From = matcher

	// parse query params and to
	// loop until we see and finished a redirect "to"
	var f string
	for {
		f, fields = takeField(fields)
		// if we got a query params
		if t := strings.Split(f, "=:"); len(t) > 1 && !strings.Contains(f, "/") {
			redirect.Queries[t[0]] = t[1]
		} else {
			redirect.To = f
			break
		}
	}

	// if there's custom status code
	var c string
	if len(fields) > 0 {
		c, fields = takeField(fields)
		if strings.HasSuffix(c, "!") {
			redirect.Shadowing = true
			c = c[0 : len(c)-1]
		}
		code, err := strconv.Atoi(c)
		if err != nil {
			return nil, fmt.Errorf("Invalid Status Code: %s", line)
		}
		redirect.StatusCode = code
	}

	// must be error if there's still something left
	if len(fields) > 0 {
		return nil, fmt.Errorf("Invalid line: %s", line)
	}

	// default status code
	if redirect.StatusCode == 0 {
		redirect.StatusCode = 301
	}

	if redirect.IsProxy() {
		// hook to all methods if it's a proxy
		for _, method := range METHODS {
			redirect.router.Handle(method, redirect.From, redirect.handler)
		}
	} else {
		redirect.router.GET(redirect.From, redirect.handler)
	}

	return &redirect, nil
}

// unshift a string because I'm lazy
func takeField(fields [][]byte) (string, [][]byte) {
	return string(fields[0]), fields[1:]
}
