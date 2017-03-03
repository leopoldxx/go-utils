package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/leopoldxx/go-utils/middleware"
)

func TestMiddlerware(t *testing.T) {
	hh := func(w http.ResponseWriter, r *http.Request) {
		t.Log("http test handler")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello middleware"))
	}

	create := func(name string, others ...interface{}) Middleware {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t.Log("http test middleware enter:", name, others)
				defer t.Log("http test middleware exit:", name, others)
				next(w, r)
			}
		}
	}

	tmw := Trace("middleware_test")
	m1 := create("middleware 1", 1, 2, 3, t)
	m2 := create("middleware 2", 4, 5, t)
	m3 := create("middleware 3")

	newhh := Chain(tmw, m1, m2, m3).HandlerFunc(hh)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	newhh(w, req)
	if w.Code != 200 || string(w.Body.Bytes()) != "hello middleware" {
		t.Fatal("fail chain handler:", w)
	}

}
