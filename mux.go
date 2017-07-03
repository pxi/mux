// Package mux implements HTTP request multiplexing helpers.
package mux

import (
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"
)

// Vars holds the named variables extracted by Match.
type Vars []struct{ k, v string }

// Match reports whether text matches the given pattern. The pattern syntax is:
//
//  pattern:
//      { term }
//  term:
//      '*'         matches any sequence of non-/ characters
//      '{' { variable-name } '}'
//                  named variable (must be non-empty); matches any sequence
//                  of non-/ characters
//      '?'         matches any single non-/ character
//      c           matches character c (c != '*', '?', '{')
//
//  variable-name:
//      c           matches character c (c != '}')
//
// Match requires pattern to match all of text, not just a substring.
// Named variables defined in the pattern are extracted to vars.
func Match(pattern, text string, vars *Vars) bool {
	key := ""
	nx, vx := 0, 0
	px, tx := 0, 0
	nextPx := 0
	nextTx := 0
	for px < len(pattern) || tx < len(text) {
		if px < len(pattern) {
			switch c := pattern[px]; c {
			default:
				if tx < len(text) && text[tx] == c {
					if px > 0 && pattern[px-1] == '}' {
						vars.Set(key, text[vx:tx])
					}
					px++
					tx++
					continue
				}
			case '?':
				if tx < len(text) && text[tx] != '/' {
					_, n := utf8.DecodeRuneInString(text[tx:])
					px += 1
					tx += n
					continue
				}
			case '*', '{':
				// Try to match at tx. If that doesn't work out,
				// restart at tx+1 next.
				nextPx = px
				nextTx = tx + 1
				px++
				if c == '{' {
					if nx < px {
						vx = tx
						nx = px + strings.IndexByte(pattern[px:], '}')
						key = pattern[px:nx]
					}
					px += len(key) + 1
				}
				continue
			}
		}
		if nextTx <= len(text) {
			px = nextPx
			tx = nextTx
			// Variable-length wildcards cannot skip /.
			if (pattern[px] == '*' || pattern[px] == '{') && text[tx-1] != '/' {
				continue
			}
		}
		vars.Reset()
		return false
	}
	if px > 0 && pattern[px-1] == '}' {
		vars.Set(key, text[vx:tx])
	}
	return true
}

// Set assigns the given value to the given key.
func (vars *Vars) Set(key, value string) {
	for i, p := range *vars {
		if p.k == key {
			(*vars)[i] = struct{ k, v string }{key, value}
			return
		}
	}
	*vars = append(*vars, struct{ k, v string }{key, value})
}

// Get returns the value for the given key. If key is not found, it returns an
// empty string.
func (vars *Vars) Get(key string) string {
	for _, p := range *vars {
		if p.k == key {
			return p.v
		}
	}
	return ""
}

// Reset resets vars to an empty state.
func (vars *Vars) Reset() { *vars = (*vars)[:0] }

// Method dispatches requests to different handlers based on the incoming
// request's HTTP method. It can automatically handle OPTIONS requests based
// on the configured handlers. This can be overridden by providing a custom
// handler for OPTIONS method. Configured HTTP method is always assumed to be
// the upper-case variant.
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
		http.Error(rw, "405 method not allowed", http.StatusMethodNotAllowed)
	}
}
