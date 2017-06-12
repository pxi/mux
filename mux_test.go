package mux

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// equal fails the test if got is not equal to want.
func equal(tb testing.TB, got, want interface{}, msg string) {
	if !reflect.DeepEqual(got, want) {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("%s:%d: "+msg+"\n got: %#v\nwant: %#v", filepath.Base(file), line, got, want)
		tb.FailNow()
	}
}

// serve returns a handler that sends a response with the given code.
func serve(code int) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(code)
	})
}

func TestMethod(t *testing.T) {
	cases := []struct {
		name    string
		methods []string // registered methods
		method  string   // request method
		code    int      // response status code
		allow   string   // response allow header
	}{
		{
			"empty",
			[]string{},
			http.MethodGet,
			http.StatusMethodNotAllowed,
			"OPTIONS",
		},
		{
			"match",
			[]string{http.MethodGet},
			http.MethodGet,
			http.StatusOK,
			"",
		},
		{
			"options",
			[]string{
				http.MethodGet,
				http.MethodPut,
				http.MethodPost,
				http.MethodPatch,
			},
			http.MethodOptions,
			http.StatusOK,
			"GET, OPTIONS, PATCH, POST, PUT",
		},
		{
			"options override",
			[]string{http.MethodOptions},
			http.MethodOptions,
			http.StatusOK,
			"",
		},
		{
			"lowercase request",
			[]string{http.MethodGet},
			"get",
			http.StatusOK,
			"",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux := Method{}
			for _, m := range tc.methods {
				mux[m] = serve(200)
			}

			rw := httptest.NewRecorder()
			req := &http.Request{Method: tc.method}

			mux.ServeHTTP(rw, req)
			equal(t, rw.Code, tc.code, "status code")
			equal(t, rw.Header().Get("Allow"), tc.allow, "allow header")
		})
	}
}
