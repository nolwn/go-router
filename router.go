package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type route struct {
	method   string
	path     string
	callback http.HandlerFunc
}

type segment struct {
	path     string
	methods  map[string]bool
	callback http.HandlerFunc
	children map[string]*segment
}

// Router is the main router object that keeps track of an looks up routes.
type Router struct {
	routes []route
	lookup *segment
}

// NewRouter is a constructor for Router.
func NewRouter() (r Router) {
	return
}

// AddRoute adds a new route with a corresponding callback to the router.
func (r *Router) AddRoute(method string, path string, callback http.HandlerFunc) (err error) {
	keys := setupKeys(strings.Split(path, "/"))
	if r.lookup == nil {
		r.lookup = &segment{}
		r.lookup.children = map[string]*segment{}
		r.lookup.methods = map[string]bool{}
	}

	curr := r.lookup

	fmt.Printf("Keys: %v", len(keys))

	for _, key := range keys {
		var seg segment
		if child, ok := curr.children[key]; !ok {
			seg = *newSegment(curr.path, key)
			curr.children[key] = &seg
			curr = &seg
		} else {
			curr = child
		}
	}

	if curr.methods[method] {
		err = errors.New("path already exists")
	}

	if err == nil {
		curr.callback = callback
		curr.methods[method] = true
		r.routes = append(r.routes, route{method, path, callback})
	}

	return
}

func addSegment(curr *segment, keys []string) (seg *segment) {
	for _, key := range keys {
		if child, ok := curr.children[key]; !ok {
			seg = newSegment(curr.path, key)
			curr.children[key] = seg
			curr = seg
		} else {
			curr = child
		}
	}

	return
}

func setupKeys(slice []string) (clean []string) {
	clean = append(clean, "/")
	for _, v := range slice {
		if v != "" {
			clean = append(clean, "/"+v)
		}
	}

	return
}

// Get is a convinience method which calls Router.AddRoute with the "GET" method.
func (r *Router) Get(path string, callback http.HandlerFunc) {
	r.AddRoute(http.MethodGet, path, callback)
}

func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	return
}

func newSegment(parentPath string, key string) (seg *segment) {
	var path string
	if parentPath == "/" {
		path = key
	} else {
		path = parentPath + key
	}
	seg = &segment{}
	seg.children = map[string]*segment{}
	seg.methods = map[string]bool{}
	seg.path = path

	return
}
