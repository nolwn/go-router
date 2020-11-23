package router

import (
	"errors"
	"net/http"
	"strings"
)

// Router is a replacement for the net/http DefaultServerMux. This version includes the
// ability to add path parameter in the given path.
//
// Paths are registered relative to their base path, WITHOUT a hostname, something that
// is allowed in the DefaultServerMux but is not allowed in this one. Each callback needs
// to be given a unique combination of method and path.
//
// Path parameters can be registered by prefacing any section of the path with a ":", so
// "/items/:itemid" would register ":itemid" as a wildcard which will be turned into  a path
// parameter called "itemid". A request path with "/items/" followed by a string of legal http
// characters, not including a slash, would match this path.
type Router struct {
	routes []route
	root   *segment

	NotFoundHandler http.Handler
}

type route struct {
	method   string
	path     string
	callback http.HandlerFunc
}

type segment struct {
	path     string
	methods  map[string]http.HandlerFunc
	children map[string]*segment
}

// NotFoundHandler is the default function for handling routes that are not found. If you wish to
// provide your own handler for this, simply set it on the router.
var NotFoundHandler http.Handler = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found."))
	})

// AddRoute registers a new handler function to a path and http.HandlerFunc. If a path and
// method already have a callback registered to them, an error is returned.
func (r *Router) AddRoute(method string, path string, callback http.HandlerFunc) (err error) {
	keys := setupKeys(strings.Split(path, "/"))

	if r.root == nil {
		r.root = &segment{}
		r.root.children = map[string]*segment{}
		r.root.methods = map[string]http.HandlerFunc{}
		r.root.path = "/"
	}

	curr := r.root

	for i, key := range keys {
		if i == 0 {
			continue
		}

		if child, ok := curr.children[key]; !ok {
			seg := newSegment(curr.path, key)
			curr.children[key] = seg
			curr = seg
		} else {
			curr = child
		}
	}

	if _, ok := curr.methods[method]; ok {
		err = errors.New("path already exists")

		return
	}

	curr.methods[method] = callback
	r.routes = append(r.routes, route{method, path, callback})

	return
}

// Get is a convinience method which calls Router.AddRoute with the "GET" method.
func (r *Router) Get(path string, callback http.HandlerFunc) {
	r.AddRoute(http.MethodGet, path, callback)
}

// Handler returns the handler to use for the given request, consulting r.Method, r.URL.Path. It
// always returns a non-nil handler.
//
// Handler also returns the registered pattern that matches the request.
//
// If there is no registered handler that applies to the request, Handler returns a ``page not
// found'' handler and an empty pattern.
func (r *Router) Handler(req *http.Request) (h http.Handler, pattern string) {
	method := req.Method
	path := req.URL.Path
	root := r.root
	curr := root

	segments := strings.Split(path, "/")
	keys := setupKeys(segments)

	if r.NotFoundHandler == nil {
		h = NotFoundHandler
	}

	for _, v := range keys {
		if v == "/" {
			continue
		}

		if seg, ok := curr.children[v]; ok {
			curr = seg
		} else {
			return
		}
	}

	if cb, ok := curr.methods[method]; ok {
		h = cb
		pattern = curr.path
	}

	return
}

// ServeHTTP is the function that is required by http.Handler. It takes an http.ResponseWriter which
// it uses to write to a response object that will construct a response for the user. It also takes
// an *http.Request which describes the request the user has made.
//
// In the case of this router, all it needs to do is lookup the Handler that has been saved at a given
// path and then call its ServeHTTP.
func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, _ := r.Handler(req)
	handler.ServeHTTP(w, req)

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

func newSegment(parentPath string, key string) (seg *segment) {
	var path string
	if parentPath == "/" {
		path = key
	} else {
		path = parentPath + key
	}
	seg = &segment{}
	seg.children = map[string]*segment{}
	seg.methods = map[string]http.HandlerFunc{}
	seg.path = path

	return
}

func setupKeys(slice []string) (keys []string) {
	keys = append(keys, "/")
	for _, v := range slice {
		if v != "" {
			keys = append(keys, "/"+v)
		}
	}

	return
}
