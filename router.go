package router

import "net/http"

type route struct {
	method   string
	path     string
	callback http.HandlerFunc
}

// Router is the main router object that keeps track of an looks up routes.
type Router struct {
	routes []route
}

// NewRouter is a constructor for Router.
func NewRouter() (r Router) {
	return
}

// AddRoute adds a new route with a corresponding callback to the router.
func (r *Router) AddRoute(method string, path string, callback http.HandlerFunc) (err error) {
	r.routes = append(r.routes, route{method, path, callback})

	return
}

// Get is a convinience method which calls Router.AddRoute with the "GET" method.
func (r *Router) Get(path string, callback http.HandlerFunc) {
	r.AddRoute(http.MethodGet, path, callback)
}

// // Handle I don't know what this is yet, I assume it's called when there's a request
// func (r *Router) Handle(w http.ResponseWriter, req *http.Request) {
// 	return
// }

func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	return
}
