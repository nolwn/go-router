package router

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SetupTestServer() (handler http.Handler) {
	// r := NewRouter()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You called the root endpoint."))
	})

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})

	return
}

func TestServer(t *testing.T) {
	var client http.Client

	handler := SetupTestServer()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	request, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	res, err := client.Do(request)

	if err != nil {
		t.Fatal("Could not make request", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if string(body) != "You called the root endpoint." {
		t.Errorf("Did not recieve the correct response. Expected 'You called the root endpoint.' Received: %s", string(body))
	}
}
