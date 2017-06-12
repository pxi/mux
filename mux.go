// Package mux implements several HTTP request multiplexers.
package mux

import (
	"net/http"
	"sort"
	"strings"
)

// Method dispatches requests to different handlers based on the incoming
// request's HTTP method. It automatically handles OPTIONS requests based on
// the configured handlers. This can be overridden by providing a custom
// handler for OPTIONS method.
type Method map[string]http.Handler

// Get dispatches GET requests to the given handler.
func Get(h http.Handler) Method { return Method{http.MethodGet: h} }

// Post dispatches POST requests to the given handler.
func Post(h http.Handler) Method { return Method{http.MethodPost: h} }

// Put dispatches PUT requests to the given handler.
func Put(h http.Handler) Method { return Method{http.MethodPut: h} }

// ServeHTTP dispatches the request to the handler whose method matches the
// request method. If handler is not found it adds the Allow header to the
// response based on the configured handlers. If the request is not an OPTIONS
// request, it also sets the response status code to 405 (Method Not Allowed).
func (route Method) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	method := strings.ToUpper(req.Method)
	if h, ok := route[method]; ok {
		h.ServeHTTP(rw, req)
		return
	}

	allow := []string{http.MethodOptions}
	for k := range route {
		k = strings.ToUpper(k)
		if k != http.MethodOptions {
			allow = append(allow, k)
		}
	}
	sort.Strings(allow)
	rw.Header().Set("Allow", strings.Join(allow, ", "))

	if method != http.MethodOptions {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
