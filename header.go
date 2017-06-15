package main

import (
	"bytes"
	"net/http"
	"regexp"

	"fmt"

	"strings"

	"crypto/subtle"

	"github.com/julienschmidt/httprouter"
)

var (
	commentLine  = regexp.MustCompile(`^\s*#`)
	comment      = regexp.MustCompile("#.+")
	leadingSpace = regexp.MustCompile(`^\s+`)
	realm        = "Please enter your username and password for this site"
)

// HeaderRouter contains routes defined in _header file.
type HeaderRouter struct {
	*httprouter.Router
}

// NewHeaderRouter creates a HeaderRouter based on _header file
func NewHeaderRouter(config []byte) (*HeaderRouter, error) {
	router := HeaderRouter{httprouter.New()}

	lines := bytes.Split(config, []byte("\n"))

	currentPath := &Path{}
	paths := make([]*Path, 0)
	for _, line := range lines {
		// if the whole line is a comment
		if commentLine.Match(line) {
			continue
		}

		// is this a new path?
		if !leadingSpace.Match(line) {
			// we alread have a path?
			if currentPath.Path != "" {
				// we're waiting for its header?
				if len(currentPath.Headers) == 0 {
					return nil, fmt.Errorf("Expect header but got a path: %s", line)
				}
				// the path is complete, push to paths
				paths = append(paths, currentPath)
			}
			// get a new path
			p := parsePath(line)
			// skip empty line
			if p == nil {
				continue
			}
			currentPath = p
			continue
		}

		// header shoud have leading space
		if !leadingSpace.Match(line) {
			return nil, fmt.Errorf("Incorrect indent: %s", line)
		}
		newPath, err := parseHeader(line, currentPath)
		if err != nil {
			return nil, err
		}
		if newPath == nil {
			continue
		}
		currentPath = newPath
	}

	if currentPath.Path != "" {
		if len(currentPath.Headers) > 0 || len(currentPath.Auths) > 0 {
			paths = append(paths, currentPath)
		} else {
			return nil, fmt.Errorf("unclosed path")
		}
	}

	for _, path := range paths {
		// use GET because we don't really care about method
		// All we care is performing lookup based on path
		router.GET(path.Path, path.Handler)
	}

	return &router, nil
}

// Path represent one URL and their additional headers
type Path struct {
	Path    string
	Headers map[string][]string
	Auths   []Auth
}

// Auth is a username/password pair which can be used to authentication with http basic auth
type Auth struct {
	Username string
	Password string
}

// Handler is a httprouter handler which is used for adding headers to response
func (path *Path) Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// if basic auth is required
	if len(path.Auths) > 0 {
		user, pass, ok := r.BasicAuth()
		login := false
		if ok {
			for _, auth := range path.Auths {
				if subtle.ConstantTimeCompare([]byte(user), []byte(auth.Username)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(auth.Password)) == 1 {
					login = true
					break
				}
			}
		}
		if !login {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized.\n"))
		}
	}
	// joining multiple header
	for k, header := range path.Headers {
		w.Header().Set(k, strings.Join(header, ", "))
	}
}

func parsePath(line []byte) *Path {
	// remove inline comment
	line = comment.ReplaceAll(line, []byte(""))
	line = bytes.Trim(line, " \t")

	if bytes.Compare(line, []byte("")) == 0 {
		return nil
	}

	if bytes.HasSuffix(line, []byte("*")) {
		line = []byte(string(line) + "splat") // lazy way to do clone + concat
	}

	return &Path{Path: string(line), Headers: make(map[string][]string), Auths: make([]Auth, 0)}
}

func parseHeader(line []byte, currentPath *Path) (*Path, error) {
	// remove inline comment
	line = comment.ReplaceAll(line, []byte(""))
	line = bytes.Trim(line, " \t")

	if bytes.Compare(line, []byte("")) == 0 {
		return nil, nil
	}

	comps := bytes.Split(line, []byte(":"))
	key := string(comps[0])
	value := strings.Trim(string(bytes.Join(comps[1:], []byte(":"))), " \t")

	if key == "Basic-Auth" {
		authPairs := strings.Split(value, " ")
		if len(currentPath.Auths) > 0 {
			return nil, fmt.Errorf("Duplicated Basic-Auth line: %s", line)
		}
		for _, pair := range authPairs {
			p := strings.Split(pair, ":")
			username := p[0]
			password := p[1]
			currentPath.Auths = append(currentPath.Auths, Auth{username, password})
		}
		return currentPath, nil
	}

	if _, ok := currentPath.Headers[key]; !ok {
		currentPath.Headers[key] = make([]string, 0)
	}
	currentPath.Headers[key] = append(currentPath.Headers[key], value)

	return currentPath, nil
}
