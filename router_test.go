package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testLevel = 1

func TestAddRouter(t *testing.T) {
	describeTests("Test AddRouter function")

	router := Router{}

	testAddRoot(router, t)
	testAddOneSegment(router, t)
	testAddManySegments(router, t)
	// TODO: add test for error when trying duplicate method + path
}

func TestServeHTTP(t *testing.T) {
	describeTests("Test ServeHTTP function")

	router := Router{}

	testMatchesRoot(router, t)
	testMatchesLongPath(router, t)
	testMatchesPathParam(router, t)
}

func addAndCheckRoute(r *Router, method string, path string, callback http.HandlerFunc) (err error) {
	routeCount := len(r.routes)

	err = r.AddRoute(method, path, callback)

	if err != nil {
		return
	}

	if len(r.routes) != routeCount+1 {
		err = fmt.Errorf("Expected there to be %d route(s), but there are %d", routeCount+1, len(r.routes))

		return
	}

	route := r.routes[len(r.routes)-1]

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
	fmt.Printf("%p { path: \"%s\", methods: %v, children: %v, parameter: %v, parameterName: \"%s\"}\n", curr, curr.path, curr.methods, curr.children, curr.parameter, curr.parameterName)

	for _, v := range curr.children {
		checkLookup(v)
	}

	if curr.parameter != nil {
		checkLookup(curr.parameter)
	}
}

func describeTests(message string) {
	fmt.Printf("%d. %s\n", testLevel, message)
	testLevel++
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

func testAddManySegments(router Router, t *testing.T) {
	defer testOutcome("add many multiple segments", t)

	err := addAndCheckRoute(&router, http.MethodDelete, "/items/thing/man/bird/horse/poop", func(http.ResponseWriter, *http.Request) {})

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodDelete, "/items/thing/man/bird/cat/poop", func(http.ResponseWriter, *http.Request) {})

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}
}

func testAddOneSegment(router Router, t *testing.T) {
	defer testOutcome("add callbacks to a single segment", t)

	err := addAndCheckRoute(&router, http.MethodPatch, "/items", func(http.ResponseWriter, *http.Request) {})

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}
}

func testAddRoot(router Router, t *testing.T) {
	defer testOutcome("add callbacks to root", t)

	err := addAndCheckRoute(&router, http.MethodGet, "/", func(http.ResponseWriter, *http.Request) {})

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}

	err = addAndCheckRoute(&router, http.MethodPost, "/", func(http.ResponseWriter, *http.Request) {})

	if err != nil {
		t.Error("The route was not correctly added to the router: ", err)
	}
}

func testMatchesLongPath(router Router, t *testing.T) {
	defer testOutcome("match long path", t)

	path := "/items/things/stuff"
	expectedBody := "I am /items/things/stuff"
	expectedCode := 200

	router.AddRoute(http.MethodGet, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedCode)
		w.Write([]byte(expectedBody))
	})

	err := matchAndCheckRoute(&router, http.MethodGet, path, expectedBody, expectedCode)

	if err != nil {
		t.Error("Did not find the expected callback handler", err)

		return
	}
}

func testMatchesPathParam(router Router, t *testing.T) {
	defer testOutcome("match path with parameter", t)

	expectedBody := "I have a path param"
	expectedCode := 200
	path := "/items/:itemid/edit"
	reqPath := "/items/this-is-an-id/edit"

	router.AddRoute(http.MethodGet, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedCode)
		w.Write([]byte(expectedBody))
	})

	err := matchAndCheckRoute(&router, http.MethodGet, reqPath, expectedBody, expectedCode)

	if err != nil {
		t.Error("Did not find the expected callback handler", err)

		return
	}
}

func testMatchesRoot(router Router, t *testing.T) {
	defer testOutcome("match root path", t)

	expectedBody := "I am /"
	expectedCode := 200
	path := "/"

	router.AddRoute(http.MethodGet, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedCode)
		w.Write([]byte(expectedBody))
	})

	err := matchAndCheckRoute(&router, http.MethodGet, path, expectedBody, expectedCode)

	if err != nil {
		t.Error("Did not find the expected callback handler", err)

		return
	}
}

func testOutcome(message string, t *testing.T) {
	var status string

	if t.Failed() {
		status = "\u001b[31mx\u001b[0m"
	} else {
		status = "\u001b[32m✓\u001b[0m"
	}

	fmt.Printf("\t%s %s\n", status, message)

	return
}
