package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/leopoldxx/go-utils/trace"
)

var defaultResponseInterceptor Recorder

// SetDefaultResponseInterceptor for http handler response
func SetDefaultResponseInterceptor(r Recorder) {
	defaultResponseInterceptor = r
}

type responseWriter struct {
	http.ResponseWriter
	sync.Mutex

	status int
	size   int
}

func (rs *responseWriter) Header() http.Header {
	return rs.ResponseWriter.Header()
}

func (rs *responseWriter) Write(data []byte) (int, error) {
	rs.Lock()
	rs.size += len(data)
	rs.Unlock()
	return rs.ResponseWriter.Write(data)
}

func (rs *responseWriter) WriteHeader(status int) {
	rs.Lock()
	rs.status = status
	rs.Unlock()
	rs.ResponseWriter.WriteHeader(status)
}

// Recorder for http handler response status & body size
type Recorder interface {
	Record(ctx context.Context, statistics Statistics)
}

// NewLogRecorder for log purpose
func NewLogRecorder() Recorder {
	return &logRecorder{}
}

type logRecorder struct{}

func (lr logRecorder) Record(ctx context.Context, statistics Statistics) {
	tracer := trace.GetTraceFromContext(ctx)
	tracer.Infof("%+v", statistics)
}

// NewMultiRecorder will chain MultiRecorder
func NewMultiRecorder(recorders ...Recorder) Recorder {
	return &multiRecorder{recorders: recorders}
}

type multiRecorder struct {
	recorders []Recorder
}

func (mr multiRecorder) Record(ctx context.Context, statistics Statistics) {
	var wg sync.WaitGroup
	for i := range mr.recorders {
		wg.Add(1)
		go func(r Recorder, ctx context.Context, statistics Statistics) {
			defer wg.Done()
			r.Record(ctx, statistics)
		}(mr.recorders[i], ctx, statistics)
	}
	wg.Wait()
}

// Statistics for http handler response
type Statistics struct {
	Status   int
	BodySize int
}

func (rs *responseWriter) Record(ctx context.Context, recorder Recorder) {
	var s Statistics
	rs.Lock()
	s.Status = rs.status
	s.BodySize = rs.size
	rs.Unlock()
	if recorder != nil {
		recorder.Record(ctx, s)
	}
}
