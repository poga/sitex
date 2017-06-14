package main

import (
	"bytes"
	"net/http"
	"regexp"

	"fmt"

	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	commentLine  = regexp.MustCompile(`^\s*#`)
	comment      = regexp.MustCompile("#.+")
	leadingSpace = regexp.MustCompile(`^\s+`)
)

// HeaderRouter contains routes defined in _header file.
type HeaderRouter struct {
	*httprouter.Router
}

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
		fmt.Printf("line: %v\n", string(line))

		// header shoud have leading space
		if !leadingSpace.Match(line) {
			return nil, fmt.Errorf("Incorrect indent: %s", line)
		}
		newPath, err := parseHeader(line, currentPath)
		if err != nil {
			return nil, fmt.Errorf("Incorrect header: %s", line)
		}
		if newPath == nil {
			continue
		}
		currentPath = newPath
	}

	if currentPath.Path != "" {
		if len(currentPath.Headers) > 0 {
			paths = append(paths, currentPath)
		} else {
			return nil, fmt.Errorf("unclosed path")
		}
	}

	for _, path := range paths {
		fmt.Printf("Defining %s\n", path.Path)
		// use GET because we don't really care about method
		// All we care is performing lookup based on path
		router.GET(path.Path, path.Handler)
	}

	return &router, nil
}

type Path struct {
	Path    string
	Headers map[string][]string
}

func (path *Path) Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// joining multiple header
	for k, header := range path.Headers {
		w.Header().Set(k, strings.Join(header, ", "))
	}
}

func parsePath(line []byte) *Path {
	// remove inline comment
	comment.ReplaceAll(line, []byte(""))
	line = bytes.Trim(line, " \t")

	if bytes.Compare(line, []byte("")) == 0 {
		return nil
	}

	if bytes.HasSuffix(line, []byte("*")) {
		line = []byte(string(line) + "splat") // lazy way to do clone + concat
	}

	return &Path{Path: string(line), Headers: make(map[string][]string)}
}

func parseHeader(line []byte, currentPath *Path) (*Path, error) {
	// remove inline comment
	comment.ReplaceAll(line, []byte(""))
	line = bytes.Trim(line, " \t")

	if bytes.Compare(line, []byte("")) == 0 {
		return nil, nil
	}

	comps := bytes.Split(line, []byte(":"))
	if len(comps) != 2 {
		return nil, fmt.Errorf("Invalid header: %s", line)
	}
	key := string(comps[0])

	if _, ok := currentPath.Headers[key]; !ok {
		currentPath.Headers[key] = make([]string, 0)
	}
	currentPath.Headers[key] = append(currentPath.Headers[key], string(bytes.Trim(comps[1], " \t")))

	return currentPath, nil
}
