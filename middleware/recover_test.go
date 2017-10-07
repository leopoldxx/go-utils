package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/leopoldxx/go-utils/middleware"
)

func TestRecover(t *testing.T) {
	ph := func(w http.ResponseWriter, r *http.Request) {
		t.Log("oops!")
		panic("oops!")
	}
	ph2 := func(w http.ResponseWriter, r *http.Request) {
		t.Log("+_+!")
		panic("+_+!")
	}

	rmw := RecoverWithTrace("rwt")

	SetDefaultResponseInterceptor(NewMultiRecorder(NewLogRecorder(), NewLogRecorder(), NewLogRecorder()))

	newhh := Chain(rmw).HandlerFunc(ph)
	newhh2 := Chain(rmw).HandlerFunc(ph2)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	newhh(w, req)
	newhh2(w, req)
	if w.Code != 500 {
		t.Fatal("fail chain handler:", w)
	}
	t.Log(len(w.Body.Bytes()), string(w.Body.Bytes()))

}
