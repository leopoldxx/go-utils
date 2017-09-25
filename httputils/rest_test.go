package httputils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leopoldxx/go-utils/trace"
)

type Req struct {
	Hello string `json:"hello"`
}

type OKResp struct {
	Message string `json:"message"`
}
type ErrResp struct {
	Error string `json:"error"`
}

func TestRest(t *testing.T) {
	testCases := []struct {
		method   string
		resource string
		req      string
		resp     string
		status   int
	}{
		{
			method:   "GET",
			resource: "/fake/get/path",
			resp:     "fake resp",
			status:   http.StatusBadRequest,
		},
		{
			method:   "POST",
			resource: "/fake/post/path",
			req:      "world!",
			resp:     "fake resp",
			status:   http.StatusOK,
		},
	}
	for _, test := range testCases {
		ts := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					t.Log(r.URL)
					w.WriteHeader(test.status)
					var body []byte
					if test.status != http.StatusOK {
						body, _ = json.Marshal(&ErrResp{Error: test.resp})
					} else {
						body, _ = json.Marshal(&OKResp{Message: test.resp})
					}
					w.Write(body)
				}))
		host := ts.URL

		msgResp := &OKResp{}
		errResp := &ErrResp{}
		ctx := trace.WithTraceForContext(context.TODO(), "rest")
		tracer := trace.GetTraceFromContext(ctx)
		tracer.Info("do request")

		resp, err := NewRestCli().
			Context(ctx).
			Method(test.method).
			Host(host).
			ResourcePath(test.resource).
			Into("2xx", msgResp).
			Into("4xx", errResp).
			Debug(Debug2).
			Do()

		if err != nil {
			t.Fatal(err)
		}
		tracer.Infof("%v", resp)

		t.Log(resp.Status)
		t.Log(resp.Header)
		t.Log(string(resp.Body))
		t.Logf("msg: %v", msgResp)
		t.Logf("err: %v", errResp)
		if test.status == http.StatusOK {
			if len(msgResp.Message) == 0 && msgResp.Message != test.resp {
				t.Fatal("expect ok resp, but failed")
			}
		} else {
			if len(errResp.Error) == 0 && errResp.Error != test.resp {
				t.Fatal("expect error resp, but failed")
			}
		}

	}
}
