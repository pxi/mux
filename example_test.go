package mux

import "net/http"

func ExampleMatch() {
	const (
		routeOne = "/{named}one/fo?/bar*"
		routeTwo = "/{named}two/fo?/bar*"
	)

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		vars := Vars{}
		path := req.URL.EscapedPath()
		switch {
		case Match(routeOne, path, &vars):
			// handle route one
		case Match(routeTwo, path, &vars):
			// handle route two
		default:
			http.NotFound(rw, req)
			return
		}
	})
}

func ExampleMethod() {
	get := func(rw http.ResponseWriter, req *http.Request) {
		// handle GET request
	}

	put := func(rw http.ResponseWriter, req *http.Request) {
		// handle PUT request
	}

	http.Handle("/", Method{
		http.MethodGet: http.HandlerFunc(get),
		http.MethodPut: http.HandlerFunc(put),
	})
}

func ExampleNotAllowed() {
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if NotAllowed(rw, req, http.MethodGet, http.MethodPut) {
			return
		}

		switch req.Method {
		case http.MethodGet:
			// handle GET request
		case http.MethodPut:
			// handle PUT request
		default:
			panic("never reached")
		}
	})
}