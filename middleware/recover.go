package middleware

import (
	"context"
	"net/http"

	"github.com/leopoldxx/go-utils/trace"
)

// RecoverWithTrace middleware is a RecoverMiddleware wraps with a trace handler
func RecoverWithTrace(name string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var rw *responseWriter
			if defaultResponseInterceptor != nil {
				rw = &responseWriter{
					ResponseWriter: w,
					status:         http.StatusOK,
				}
			}
			recoverHandler := func(w http.ResponseWriter, r *http.Request) {
				tracer := trace.GetTraceFromRequest(r)
				if rw, ok := w.(interface {
					Record(ctx context.Context, recorder Recorder)
				}); ok {
					defer rw.Record(r.Context(), defaultResponseInterceptor)
				}
				defer func() {
					if err := recover(); err != nil {
						tracer.Error("panic:", tracer.Stack())
						http.Error(w, "internal panic error, plz check log!", http.StatusInternalServerError)
					}
				}()
				next(w, r)
			}
			if rw != nil {
				trace.HandleFunc(name, recoverHandler)(rw, r)
				return
			}

			trace.HandleFunc(name, recoverHandler)(w, r)
		}
	}
}
