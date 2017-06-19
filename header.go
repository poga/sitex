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

// Header contains routes defined in _header file.
type Header struct {
	router *httprouter.Router
}

func (header Header) Match(r *http.Request) bool {
	handle, _, _ := header.router.Lookup(r.Method, r.URL.Path)
	return handle != nil
}

// Handle checks if the router should process the given request.
// It's a noop if the router should not process the request.
// Returns error if the response is finalized and we shouldn't return anything more.
func (header Header) Handle(w http.ResponseWriter, r *http.Request) (bool, error) {
	if !header.Match(r) {
		return true, nil
	}
	handle, params, _ := header.router.Lookup("GET", r.URL.Path)
	if handle != nil {
		handle(w, r, params)

		// if there's an authentication error. stop the handler chain
		if w.Header().Get("WWW-Authenticate") != "" {
			return false, fmt.Errorf("Unauthorized")
		}
	}
	return true, nil
}

// NewHeaders returns an list of HeaderRouters from given rules.
// Every path will creates a HeaderRouter
func NewHeaders(config []byte) ([]Header, error) {
	headers := make([]Header, 0)

	lines := bytes.Split(config, []byte("\n"))

	currentPath := &path{}
	header := Header{router: httprouter.New()}
	for _, line := range lines {
		// skip comment line
		if commentLine.Match(line) {
			continue
		}
		// skip empty line
		if bytes.Compare(bytes.Trim(line, " \t"), []byte("")) == 0 {
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
				header.router.GET(currentPath.Path, currentPath.Handler)
				headers = append(headers, header)
				header = Header{router: httprouter.New()}
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

		// are we waiting for header?
		if currentPath.Path == "" {
			return nil, fmt.Errorf("Dangling header without path: %s', line")
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
			header.router.GET(currentPath.Path, currentPath.Handler)
			headers = append(headers, header)
		} else {
			return nil, fmt.Errorf("unclosed path")
		}
	}

	return headers, nil
}

type path struct {
	Path    string
	Headers map[string][]string
	Auths   []auth
}

type auth struct {
	Username string
	Password string
}

// Handler is a httprouter handler which is used for adding headers to response
func (path *path) Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func parsePath(line []byte) *path {
	// remove inline comment
	line = comment.ReplaceAll(line, []byte(""))
	line = bytes.Trim(line, " \t")

	if bytes.Compare(line, []byte("")) == 0 {
		return nil
	}

	if bytes.HasSuffix(line, []byte("*")) {
		line = []byte(string(line) + "splat") // lazy way to do clone + concat
	}

	return &path{Path: string(line), Headers: make(map[string][]string), Auths: make([]auth, 0)}
}

func parseHeader(line []byte, currentPath *path) (*path, error) {
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
			currentPath.Auths = append(currentPath.Auths, auth{username, password})
		}
		return currentPath, nil
	}

	if _, ok := currentPath.Headers[key]; !ok {
		currentPath.Headers[key] = make([]string, 0)
	}
	currentPath.Headers[key] = append(currentPath.Headers[key], value)

	return currentPath, nil
}
