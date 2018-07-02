package server

import (
	"errors"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/facebookgo/httpdown"
	"github.com/leopoldxx/go-utils/trace/glog"

	"github.com/gorilla/mux"
)

// Controller is an interface that register handlers with a http router
type Controller interface {
	Register(router *mux.Router)
}

// Server is the api server interface
type Server interface {
	http.Handler
	Register(ctrl Controller)
	ListenAndServe() error
}

type options struct {
	listenAddr      string
	prefix          string
	debug           bool
	notfoundHandler http.Handler
}

// Option func for server
type Option func(opts *options)

// ListenAddr will set listen addr for the server
func ListenAddr(addr string) Option {
	return func(opts *options) {
		opts.listenAddr = addr
	}
}

// APIPrefix will set api prefix for the api server
func APIPrefix(prefix string) Option {
	return func(opts *options) {
		opts.prefix = prefix
	}
}

// PProf switch on/off the http pprof api
func PProf(d bool) Option {
	return func(opts *options) {
		opts.debug = d
	}
}

// WithNotFoundHandler set NotFoundHandler for router
func WithNotFoundHandler(h http.Handler) Option {
	return func(opts *options) {
		opts.notfoundHandler = h
	}
}

func debug(router *mux.Router) {
	router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
}

type server struct {
	listenAddr string
	prefix     string
	rrouter    *mux.Router
	router     *mux.Router
}

// New func for server creating
func New(ops ...Option) Server {
	opts := &options{
		listenAddr: ":8080", // default listen port
	}
	for idx := range ops {
		ops[idx](opts)
	}

	s := &server{
		listenAddr: opts.listenAddr,
		prefix:     opts.prefix,
		rrouter:    mux.NewRouter(),
	}

	if opts.debug == true {
		debug(s.rrouter)
	}

	if opts.notfoundHandler != nil {
		s.rrouter.NotFoundHandler = opts.notfoundHandler
	}
	s.router = s.rrouter
	if len(s.prefix) != 0 {
		s.router = s.rrouter.PathPrefix(s.prefix).Subrouter()
	}

	return s
}

func (s *server) Register(ctrl Controller) {
	if s == nil || ctrl == nil {
		return
	}
	ctrl.Register(s.router)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s == nil {
		panic("nil server")
	}
	s.rrouter.ServeHTTP(w, r)
}

func (s *server) ListenAndServe() error {
	if s == nil {
		return errors.New("nil server")
	}
	httpServer := &http.Server{
		Addr:    s.listenAddr,
		Handler: s.rrouter,
	}

	hd := &httpdown.HTTP{
		StopTimeout: time.Second,
		KillTimeout: time.Second,
	}

	glog.Infof("HTTP server listening on %s", s.listenAddr)
	defer glog.Flush()
	defer glog.Info("HTTP server stopped")

	if err := httpdown.ListenAndServe(httpServer, hd); err != nil {
		glog.Errorf("listen and serve failed: %s", err)
		return err
	}
	return nil
}
