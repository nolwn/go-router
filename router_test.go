package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddRouter(t *testing.T) {
	router := Router{}
	routeCounter := 0

	err := addAndCheckRoute(&router, http.MethodGet, "/", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodPost, "/", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodPatch, "/items", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodDelete, "/items/thing/man/bird/horse/poop", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodDelete, "/items/thing/man/bird/cat/poop", func(http.ResponseWriter, *http.Request) {}, &routeCounter)

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	// checkLookup(router.lookup)
}

func TestServeHTTP(t *testing.T) {
	router := Router{}
	path := "/items"
	expectedBody := "I am /items"
	expectedCode := 200

	router.AddRoute(http.MethodGet, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedCode)
		w.Write([]byte(expectedBody))
	})

	err := matchAndCheckRoute(&router, http.MethodGet, path, expectedBody, expectedCode)

	if err != nil {
		t.Error("Did not find the expected callback handler", err)
	}
}

func matchAndCheckRoute(r *Router, method string, path string, expectedBody string, expectedCode int) (err error) {
	request, err := http.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()

	if err != nil {
		err = fmt.Errorf("Could not create request")

		return
	}

	r.ServeHTTP(rr, request)

	if rr.Code != expectedCode {
		err = fmt.Errorf("The returned callback did not write 200 to the header. Found %d", rr.Code)

		return
	}

	body, _ := ioutil.ReadAll(rr.Body)

	if string(body) != string([]byte(expectedBody)) {
		err = fmt.Errorf(
			"The returned callback did not write the expected body. Expected: %s. Actual: %s",
			expectedBody,
			string(body),
		)

		return
	}

	return
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

// checkLookup prints out the various saved routes. It's not needed for any test, but is a helpful debugging tool.
func checkLookup(curr *segment) {
	fmt.Printf("%p { path: %s, methods: %v, children: %v}\n", curr, curr.path, curr.methods, curr.children)

	for _, v := range curr.children {
		checkLookup(v)
	}
}
