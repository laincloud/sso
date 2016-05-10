/*
Package server provides basic wrapping for go web server. Supporting:
  * Context derived from base context for each request
  * Named routing and url reversing
  * Middleware supports
  * Server graceful shutdown
  * Restful resource for json restful api
*/
package server

import (
	"html/template"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mijia/sweb/log"
	"github.com/stretchr/graceful"
	"golang.org/x/net/context"
)

// Handler is an function signature to be registered to serve routing requests with context support.
// Server would inject the context to the request handler.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context

const (
	kHrParamsKey     = "inter_ctx_key_hrparams"
	kGracefulTimeout = 10
)

type Muxer interface {
	Middleware(ware Middleware)
	Handle(method, path, name string, handle Handler)
	Get(path, name string, handle Handler)
	Post(path, name string, handle Handler)
	Put(path, name string, handle Handler)
	Patch(path, name string, handle Handler)
	Delete(path, name string, handle Handler)
	Head(path, name string, handle Handler)
}

// Server is a struct for all kinds of internal data.
type Server struct {
	srv                *graceful.Server
	baseCtx            context.Context
	wares              []Middleware
	router             *httprouter.Router
	extraAssetsMapping map[string]string
	extraAssetsJson    string
	assetsPrefix       string
	namedRoutes        map[string]string
	restfulAdapter     RestfulHandlerAdapter
	debug              bool
}

// Run the server listen and server at the addr with graceful shutdown supports.
func (s *Server) Run(addr string) error {
	timeout := kGracefulTimeout * time.Second
	if s.debug {
		timeout = 0
	}
	s.srv = &graceful.Server{
		Timeout: timeout,
		Server: &http.Server{
			Addr:    addr,
			Handler: s.router,
		},
	}
	log.Infof("Server is listening on %s", addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Stop(timeout time.Duration) {
	s.srv.Stop(timeout)
}

// Middleware: Register a middleware to a server object.
func (s *Server) Middleware(ware Middleware) {
	s.wares = append(s.wares, ware)
}

// Handle: basic interface which register a http request and handler to the router
func (s *Server) Handle(method, path, name string, handle Handler) {
	s.router.Handle(method, path, s.hrAdapt(handle))
	s.namedRoutes[name] = path
}

// Get will register a 'GET' request handler to the router.
func (s *Server) Get(path string, name string, handle Handler) {
	s.Handle("GET", path, name, handle)
}

// Post will register a 'POST' request handler to the router.
func (s *Server) Post(path string, name string, handle Handler) {
	s.Handle("POST", path, name, handle)
}

// Put will register a 'PUT' request handler to the router.
func (s *Server) Put(path string, name string, handle Handler) {
	s.Handle("PUT", path, name, handle)
}

// Patch will register a 'PATCH' request handler to the router.
func (s *Server) Patch(path string, name string, handle Handler) {
	s.Handle("PATCH", path, name, handle)
}

// Head will register a 'HEAD' request handler to the router.
func (s *Server) Head(path string, name string, handle Handler) {
	s.Handle("HEAD", path, name, handle)
}

// Delete will register a 'DELETE' request handler to the router.
func (s *Server) Delete(path string, name string, handle Handler) {
	s.Handle("DELETE", path, name, handle)
}

// NotFound wil register a 404 NotFound handler to the router.
func (s *Server) NotFound(handle Handler) {
	if handle != nil {
		h := s.hrAdapt(handle)
		s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(w, r, nil)
		})
	}
}

// MethodNotAllowed will register a 405 handler to the router
func (s *Server) MethodNotAllowed(handle Handler) {
	if handle != nil {
		h := s.hrAdapt(handle)
		s.router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(w, r, nil)
		})
	}
}

// DefaultRouteFuncs provides a FuncMap for the renderer includes 'assets' and 'urlReverse'
// so that you can use those functions inside the templates.
func (s *Server) DefaultRouteFuncs() template.FuncMap {
	return template.FuncMap{
		"assets": func(path string) (string, error) {
			return s.Assets(path), nil
		},
		"urlReverse": func(name string, params ...interface{}) (string, error) {
			return s.Reverse(name, params...), nil
		},
	}
}

// htAdapt adapts a sweb Handler to the httprouter Handle
func (s *Server) hrAdapt(fn Handler) httprouter.Handle {
	core := func(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
		// we are inside the onion core, so the next would be ignored
		if s.debug {
			s.loadJsonAssetsMapping()
		}
		return fn(ctx, w, r)
	}
	handler := buildOnion(append(s.wares, MiddleFn(core)))
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := s.baseCtx
		if len(params) > 0 {
			ctx = newContextWithParams(ctx, params)
		}
		handler.ServeHTTP(ctx, NewResponseWriter(w), r)
	}
}

// Params extracts the param from url, e.g. "/hello/:name" -> server.Params(ctx, "name")
func Params(ctx context.Context, key string) string {
	if params, ok := ctx.Value(kHrParamsKey).(httprouter.Params); !ok {
		return ""
	} else {
		return params.ByName(key)
	}
}

// New a go web server with context as parent context
func New(ctx context.Context, isDebug bool) *Server {
	if isDebug {
		log.EnableDebug()
	}
	srv := &Server{
		baseCtx:            ctx,
		wares:              []Middleware{},
		router:             httprouter.New(),
		extraAssetsMapping: make(map[string]string),
		namedRoutes:        make(map[string]string),
		debug:              isDebug,
	}
	return srv
}

// newContextWithParams just injects the httprouter params into the context
func newContextWithParams(ctx context.Context, params httprouter.Params) context.Context {
	return context.WithValue(ctx, kHrParamsKey, params)
}
