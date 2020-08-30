package router

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAddRouter(t *testing.T) {
	r := Router{}

	err := addAndCheckRoute(r, http.MethodGet, "/", func(http.ResponseWriter, *http.Request) {}, 0)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}
}

func addAndCheckRoute(r Router, method string, path string, callback http.HandlerFunc, expectedIndex int) (err error) {
	err = r.AddRoute(method, path, callback)

	if err != nil {
		return
	}

	if len(r.routes) != expectedIndex+1 {
		err = fmt.Errorf("Expected there to be %d route(s), but there are %d", expectedIndex+1, len(r.routes))

		return
	}

	route := r.routes[expectedIndex]

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
