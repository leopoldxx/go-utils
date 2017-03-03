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

package trace_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leopoldxx/go-utils/trace"
)

func TestTrace(t *testing.T) {
	t1 := trace.New("t1")
	t2 := trace.WithParent(t1, "t2")

	t1.Info()
	t2.Info()
	t1.Info("hello trace")
	t2.Info("hello %s", "golang")
	t2.Infof("hello %s", "golang")

	t3 := trace.New("t3")
	t3.Infof("new log id, %s", "look")

	t1.Warn(t3)
	t2.Error(t3)

	t.Log(t1)
	t.Log(t2)
	t.Log(t3)
	t.Log(t1.Stack())
}
func TestTraceHandler(t *testing.T) {
	ts := httptest.NewServer(trace.Handler("httptest", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tracer trace.Trace
		if tracer = trace.GetTraceFromRequest(r); tracer == nil {
			tracer = trace.New("internal-test")
		}
		tracer.Info("process start...")
		defer tracer.Info("process end...")

		tracer.Info("hello test!")
		fmt.Fprintln(w, `hello test!`)
	})))

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal("get url failed:", err)
	}
	defer res.Body.Close()

	ts.Close()
}
