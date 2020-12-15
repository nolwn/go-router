package router

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

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
	NotFoundHandler http.Handler
	root            *segment
	routes          []route
}

// endpoint is comes at the end of each valid path in the tree. It contains the information you
// need to call the endpoint, including path parameter names.
type endpoint struct {
	callback   http.HandlerFunc
	path       string
	pathParams []string
}

// parameter contains a pointer to a parameter segment and the name of the parameter.
type parameter struct {
	name    string
	segment *segment
}

// route is not part of the tree, but is saved on the router to represent all the available
// routes in the tree.
type route struct {
	callback http.HandlerFunc
	method   string
	path     string
}

// segment is a tree node. It can have children, or endpoints, or both attached to it. It also
// has a special child called "parameter" which represents a path parameter. If a route string
// doesn't match any of the children, and there is a parameter child present, it will match that
// parameter child.
type segment struct {
	children  map[string]*segment
	endpoints map[string]*endpoint
	parameter parameter
}

var paramsKey = contextKey("params")

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
	pathParams := []string{}

	if r.root == nil {
		r.root = &segment{}
		r.root.children = map[string]*segment{}
		r.root.endpoints = map[string]*endpoint{}
	}

	curr := r.root

	for i, key := range keys {
		if i == 0 {
			continue
		}

		if isParameter(key) {
			pathParams = append(pathParams, key[2:])

		}

		if child, _ := getChild(key, curr); child == nil {
			seg := addSegment(curr, key)
			curr = seg
		} else {
			curr = child
		}
	}

	if _, ok := curr.endpoints[method]; ok {
		err = errors.New("path already exists")

		return
	}

	curr.endpoints[method] = &endpoint{callback, path, pathParams}
	r.routes = append(r.routes, route{callback, method, path})

	return
}

// ServeHTTP is the function that is required by http.Handler. It takes an http.ResponseWriter which
// it uses to write to a response object that will construct a response for the user. It also takes
// an *http.Request which describes the request the user has made.
//
// In the case of this router, all it needs to do is lookup the Handler that has been saved at a given
// path and then call its ServeHTTP.
func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path
	var handler http.Handler

	endpoint, params, err := r.getEndpoint(method, path)

	if err != nil {
		handler = r.NotFoundHandler
	} else {
		handler = endpoint.callback
		ctx := context.WithValue(context.Background(), paramsKey, params)
		req = req.WithContext(ctx)
	}

	handler.ServeHTTP(w, req)

	return
}

// PathParams takes a path and returns the values for any path parameters
// in the path.
func PathParams(req *http.Request) (params map[string]string) {
	params = req.Context().Value(paramsKey).(map[string]string)

	return
}

// addSegment create a new segment either as a child or as a parameter depending on whether the key
// qualifies as a parameter. A pointer to the created segment is then returned.
func addSegment(curr *segment, key string) (seg *segment) {
	if curr.parameter.segment != nil {
		seg = curr.parameter.segment

	} else if child, ok := curr.children[key]; !ok { // child does not match...
		var isParam bool

		seg, isParam = newSegment(key)

		if isParam {
			curr.parameter.segment = seg
			curr.parameter.name = key[2:]

		} else {
			curr.children[key] = seg
		}

		return

	} else { // child matches...
		seg = child
	}

	return
}

// getChild takes a path part and finds the appropriate segment child for it. If it is an exact match to a
// child on the segment, then that child segment is returned. If it is not a match, then the parameter child
// is returned. If there is no parameter child, nil is returned. isParam is true if the parameter child is
// being returned.
func getChild(key string, curr *segment) (child *segment, param string) {
	if seg, ok := curr.children[key]; ok { // is there an exact match?
		child = seg

	} else if curr.parameter.segment != nil { // could this be a parameter?
		child = curr.parameter.segment
		param = curr.parameter.name
	}

	return
}

// getEndpoint takes a path and traverses the tree until it finds the endpoint associated with that path.
// If no endpoint if found, an error is returned.
func (r *Router) getEndpoint(method string, path string) (end *endpoint, params map[string]string, err error) {
	curr := r.root
	segments := strings.Split(path, "/")
	params = map[string]string{}
	keys := setupKeys(segments)

	for _, key := range keys {
		if key == "/" {
			continue
		}

		seg, paramName := getChild(key, curr)

		if seg == nil {
			err = errors.New("route not found")

			return
		}

		if paramName != "" {
			params[paramName] = key[1:]
		}

		curr = seg
	}

	if _, ok := curr.endpoints[method]; !ok {
		err = errors.New("route not found")
	}

	end = curr.endpoints[method]

	return
}

// TODO: refactor out newSegment as it's not longer needed.

// newSegment constructs a new, empty segment and reports back if the key is a parameter.
func newSegment(key string) (seg *segment, isParam bool) {
	seg = &segment{}

	seg.children = map[string]*segment{}
	seg.endpoints = map[string]*endpoint{}

	if isParameter(key) {
		isParam = true
	}

	return
}

// setupKeys takes an array of strings representing the parts of a path, and returns a new slice
// made up of the parts with "/" prepended to each.
func setupKeys(slice []string) (keys []string) {
	keys = append(keys, "/")
	for _, v := range slice {
		if v != "" {
			keys = append(keys, "/"+v)
		}
	}

	return
}

// isParameter returns true if the key is more than one character long and starts with a ':'
func isParameter(key string) (isParam bool) {
	if len([]rune(key)) <= 1 {
		return // avoid empty variables, i.e. /somepath/:/someotherpath
	}

	if key[1] != ':' {
		return
	}

	isParam = true

	return
}
