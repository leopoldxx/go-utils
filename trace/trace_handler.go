// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"context"
	"net/http"
)

const (
	tracerLogHandlerID = "tracer-handler-id-757b345cf9312183e788faaee990d349"
)

// Handler wrap a trace handler outer the original http.Handler
func Handler(name string, handler http.Handler) http.Handler {
	return http.HandlerFunc(HandleFunc(name, handler.ServeHTTP))
}

// HandleFunc wrap a trace handle func outer the original http handle func
func HandleFunc(name string, handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var tracer Trace

		if id := r.Header.Get("x-request-id"); len(id) > 0 {
			tracer = WithID(name, id)
		} else {
			tracer = New(name)
		}

		realIP := func(r *http.Request) string {
			if ip, exists := r.Header["X-Real-IP"]; exists && len(ip) > 0 {
				return ip[0]
			}
			if ips, exists := r.Header["X-Forwarded-For"]; exists && len(ips) > 0 {
				return ips[0]
			}
			return r.RemoteAddr
		}

		tracer.Infof("event=[request-in] remote=[%s] method=[%s] url=[%s]", realIP(r), r.Method, r.URL.String())
		defer tracer.Info("event=[request-out]")

		w.Header().Set("x-request-id", tracer.ID())
		handler(w, r.WithContext(context.WithValue(r.Context(), tracerLogHandlerID, tracer)))
	}
}

// GetTraceFromRequest get the Trace var from the req context, if there is no such a trace utility, return nil
func GetTraceFromRequest(r *http.Request) Trace {
	return GetTraceFromContext(r.Context())
}

// GetTraceFromContext get the Trace var from the context, if there is no such a trace utility, return nil
func GetTraceFromContext(ctx context.Context) Trace {
	if tracer, ok := ctx.Value(tracerLogHandlerID).(Trace); ok {
		return tracer
	}
	return New("default-trace")
}

// WithTraceForContext will return a new context wrapped a trace handler around the original ctx
func WithTraceForContext(ctx context.Context, traceName string, traceID ...string) context.Context {
	return context.WithValue(ctx, tracerLogHandlerID, New(traceName, traceID...))
}
