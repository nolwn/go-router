package router

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAddRouter(t *testing.T) {
	r := Router{}
	routeCounter := 0

	err := addAndCheckRoute(&r, http.MethodGet, "/", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&r, http.MethodPost, "/", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&r, http.MethodPatch, "/items", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}
}

func addAndCheckRoute(r *Router, method string, path string, callback http.HandlerFunc, routeCounter *int) (err error) {
	err = r.AddRoute(method, path, callback)

	defer func(routeCounter *int) {
		*routeCounter++
	}(routeCounter)

	if err != nil {
		return
	}

	if len(r.routes) != *routeCounter+1 {
		err = fmt.Errorf("Expected there to be %d route(s), but there are %d", *routeCounter+1, len(r.routes))

		return
	}

	route := r.routes[*routeCounter]

	if route.method != method {
		err = fmt.Errorf("Expected the route method to be %s, but it was %s", method, route.method)

		return
	}

	if route.path != path {
		err = fmt.Errorf("Expected the route path to be %s, but it was %s", path, route.path)

		return
	}

	if route.callback == nil {
		err = fmt.Errorf("Expected route to have a callback function, but the callback was nil")

		return
	}

	return
}

// func TestHandle(t *testing.T) {
// 	r := NewRouter()

// 	request, _ := http.NewRequest(http.MethodGet, "http://example.domain/api", nil)
// 	var writer http.ResponseWriter

// r.Handle(writer, request)
// }
