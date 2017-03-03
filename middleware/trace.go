package middleware

import (
	"net/http"

	"github.com/leopoldxx/go-utils/trace"
)

// Trace will create a trace handler middleware
func Trace(name string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return trace.HandleFunc(name, next)
	}
}
