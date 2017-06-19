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

func TestMatch(t *testing.T) {
	none := Vars{}

	cases := []struct {
		pattern, text string
		match         bool
		vars          Vars
	}{
		{"abc", "abc", true, none},
		{"*", "abc", true, none},
		{"{1}", "abc", true, Vars{{"1", "abc"}}},
		{"*c", "abc", true, none},
		{"{1}c", "abc", true, Vars{{"1", "ab"}}},
		{"a*", "a", true, none},
		{"a*", "abc", true, none},
		{"a*", "ab/c", false, none},
		{"a{1}", "a", true, Vars{{"1", ""}}},
		{"a{1}", "abc", true, Vars{{"1", "bc"}}},
		{"a{1}", "ab/c", false, none},
		{"a*/b", "abc/b", true, none},
		{"a*/b", "a/c/b", false, none},
		{"a{1}/b", "abc/b", true, Vars{{"1", "bc"}}},
		{"a{1}/b", "a/c/b", false, none},
		{"a*b*c*d*e*/f", "axbxcxdxe/f", true, none},
		{"a*b*c*d*e*/f", "axbxcxdxexxx/f", true, none},
		{"a*b*c*d*e*/f", "axbxcxdxe/xxx/f", false, none},
		{"a*b*c*d*e*/f", "axbxcxdxexxx/fff", false, none},
		{
			"a{1}b{2}c{3}d{4}e{5}/f",
			"axbxcxdxe/f",
			true,
			Vars{{"1", "x"}, {"2", "x"}, {"3", "x"}, {"4", "x"}, {"5", ""}},
		},
		{
			"a{1}b{2}c{3}d{4}e{5}/f",
			"axbxcxdxexxx/f",
			true,
			Vars{{"1", "x"}, {"2", "x"}, {"3", "x"}, {"4", "x"}, {"5", "xxx"}},
		},
		{"a{1}b{2}c{3}d{4}e{5}/f", "axbxcxdxe/xxx/f", false, none},
		{"a{1}b{2}c{3}d{4}e{5}/f", "axbxcxdxexxx/fff", false, none},
		{"a*b?c*x", "abxbbxdbxebxczzx", true, none},
		{"a*b?c*x", "abxbbxdbxebxczzy", false, none},
		{"a{1}b?c{2}x", "abxbbxdbxebxczzx", true, Vars{{"1", "bxbbxdbxe"}, {"2", "zz"}}},
		{"a{1}b?c{2}x", "abxbbxdbxebxczzy", false, none},
		{"a?b", "a☺b", true, none},
		{"a???b", "a☺b", false, none},
		{"a?b", "a/b", false, none},
		{"a*b", "a/b", false, none},
		{"a{1}b", "a/b", false, none},
		{"*x", "xxx", true, none},
		{"{1}x", "xxx", true, Vars{{"1", "xx"}}},
		{"*/a", "/a", true, none},
		{"{1}/a", "/a", true, Vars{{"1", ""}}},
		{"/*", "/", true, none},
		{"/{1}", "/", true, Vars{{"1", ""}}},
	}

	for _, tt := range cases {
		vs := Vars{}
		ok := Match(tt.pattern, tt.text, &vs)
		if ok != tt.match {
			t.Errorf("Match(%#q, %#q) = %v, want %v", tt.pattern, tt.text, ok, tt.match)
		}
		if !reflect.DeepEqual(vs, tt.vars) {
			t.Errorf("Match(%#q, %#q)\n got: %#v\nwant: %#v", tt.pattern, tt.text, vs, tt.vars)
		}
	}
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
