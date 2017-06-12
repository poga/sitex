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

type Route struct {
	Match      string
	Queries    map[string]string
	StatusCode int
	To         string
	wd         string
}

func (route *Route) CompileRedirectTo(r *http.Request, ps httprouter.Params) string {
	var pattern = route.To
	// is there any splat in the pattern?
	if strings.HasSuffix(pattern, ":splat") {
		splat := ps.ByName("splat")
		splat = splat[1:len(splat)] // remove "/" at the beginning
		return strings.Replace(pattern, ":splat", splat, 1)
	}

	// does this require query param matching?
	if len(route.Queries) > 0 {
		result := pattern
		for query, varName := range route.Queries {
			result = strings.Replace(result, fmt.Sprintf(":%s", varName), r.URL.Query().Get(query), 1)
		}
		return result
	}

	// is there any placeholder in the pattern?
	if strings.Contains(pattern, ":") {
		varName := regexp.MustCompile(":[^/]+")
		vars := varName.FindAllString(pattern, -1)
		result := pattern
		for _, v := range vars {
			name := v[1:len(v)]
			result = strings.Replace(result, v, ps.ByName(name), 1)
		}
		return result
	}

	// no replace needed
	return pattern
}

func (route *Route) Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// if there's queries to match
	if len(route.Queries) > 0 {
		queryMatched := true
		for _, q := range route.Queries {
			if r.URL.Query().Get(q) == "" {
				queryMatched = false
				break
			}
		}
		if queryMatched {
			route.statusCodeHandler(w, r, ps)
		} else {
			http.ServeFile(w, r, filepath.Join(route.wd, r.URL.Path))
		}
		return
	}

	route.statusCodeHandler(w, r, ps)
	return
}

func (route *Route) statusCodeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if route.StatusCode >= 300 && route.StatusCode < 400 {
		http.Redirect(w, r, route.CompileRedirectTo(r, ps), route.StatusCode)
		return
	}

	if strings.HasPrefix(route.To, "http") {
		req, err := http.NewRequest(r.Method, route.To, r.Body)
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

	}

	http.ServeFile(w, r, filepath.Join(route.wd, route.CompileRedirectTo(r, ps)))
}

// ParseRedirectRule parse a line based on netlify's redirect rule format,
func ParseRedirectRule(wd string, line []byte) (*Route, error) {
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

	route := Route{Queries: make(map[string]string), wd: wd}

	// = parse match
	matcher, fields := takeField(fields)
	// if it's a splat route, add a variable name for httprouter
	if strings.HasSuffix(matcher, "*") {
		matcher = matcher + "splat"
	}
	route.Match = matcher

	// = parse query params and to
	// loop until we see and finished a redirect "to"
	var f string
	for {
		f, fields = takeField(fields)
		// if we got a query params
		if t := strings.Split(f, "=:"); len(t) > 1 && !strings.Contains(f, "/") {
			route.Queries[t[0]] = t[1]
		} else {
			route.To = f
			break
		}
	}

	// if there's custom status code
	var c string
	if len(fields) > 0 {
		c, fields = takeField(fields)
		code, err := strconv.Atoi(c)
		if err != nil {
			return nil, fmt.Errorf("Invalid Status Code: %s", line)
		}
		route.StatusCode = code
	}

	// must be error if there's still something left
	if len(fields) > 0 {
		return nil, fmt.Errorf("Invalid line: %s", line)
	}

	// default status code
	if route.StatusCode == 0 {
		route.StatusCode = 301
	}

	return &route, nil
}

// unshift a string because I'm lazy
func takeField(fields [][]byte) (string, [][]byte) {
	return string(fields[0]), fields[1:]
}
