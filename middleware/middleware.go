package middleware

import "net/http"

// Middleware is a chainable preprocessor for Endpoint
type Middleware func(http.HandlerFunc) http.HandlerFunc

// HandlerFunc will return the HandlerFunc of the middleware
func (m Middleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return m(next)
}

// Chain is a helper function for composing middlewares.
func Chain(first Middleware, others ...Middleware) Middleware {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		for i := len(others) - 1; i >= 0; i-- {
			handler = others[i](handler)
		}
		return first(handler)
	}
}
