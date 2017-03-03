package middleware

import (
	"net/http"

	"github.com/leopoldxx/go-utils/trace"
)

// Recover middleware will response a default 500 error
// when panic occurred inside of the handler process.
// NOTIC: if there is no a trace handler out the recover handler,
// prepend a default trace handler
func Recover() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tracer := trace.GetTraceFromRequest(r)
			defer func() {
				if err := recover(); err != nil {
					tracer.Error("panic:", tracer.Stack())
					http.Error(w, "internal error, plz check log!", http.StatusInternalServerError)
				}
			}()
			next(w, r)
		}
	}
}

// RecoverWithTrace middleware is a RecoverMiddleware wraps with a trace handler
func RecoverWithTrace(name string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			recoverHandler := func(w http.ResponseWriter, r *http.Request) {
				tracer := trace.GetTraceFromRequest(r)
				defer func() {
					if err := recover(); err != nil {
						tracer.Error("panic:", tracer.Stack())
						http.Error(w, "internal error, plz check log!", http.StatusInternalServerError)
					}
				}()
				next(w, r)
			}

			trace.HandleFunc(name, recoverHandler)(w, r)
		}
	}
}
