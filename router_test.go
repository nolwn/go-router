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

func TestHandler(t *testing.T) {
	router := Router{}
	request, err := http.NewRequest(http.MethodGet, "http://example.com/items", nil)
	rr := httptest.NewRecorder()
	expectedBody := "I am /items"

	if err != nil {
		t.Error("Could not create request")
	}

	router.AddRoute(http.MethodGet, "/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(expectedBody))
	})

	checkLookup(router.lookup)

	h, pattern := router.Handler(request)

	if pattern != "/items" {
		t.Errorf("The recovered patter does not match: %s", pattern)
	}

	h.ServeHTTP(rr, request)

	if rr.Code != 200 {
		t.Errorf("The returned callback did not write 200 to the header. Found %d", rr.Code)
	}

	body, _ := ioutil.ReadAll(rr.Body)

	if string(body) != string([]byte(expectedBody)) {
		t.Errorf(
			"The returned callback did not write the expected body. Expected: %s. Actual: %s",
			expectedBody,
			string(body),
		)
	}
}

func checkLookup(curr *segment) {
	fmt.Printf("%p { path: %s, methods: %v, children: %v}\n", curr, curr.path, curr.methods, curr.children)

	for _, v := range curr.children {
		checkLookup(v)
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
