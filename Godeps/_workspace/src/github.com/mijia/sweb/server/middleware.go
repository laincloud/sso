package server

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

// Middleware is an interface defining the middleware for the server, the middleware should call the next handler to pass the
// request down, or just return a HttpRedirect request and etc.
type Middleware interface {
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context
}

// MiddleFn is an adapter to adapt a function to a Middleware interface
type MiddleFn func(context.Context, http.ResponseWriter, *http.Request, Handler) context.Context

// ServeHTTP adapts the Middleware interface.
func (m MiddleFn) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
	return m(ctx, w, r, next)
}

// RecoveryWare is the recovery middleware which can cover the panic situation.
type RecoveryWare struct {
	printStack bool
	stackAll   bool
	stackSize  int
}

// ServeHTTP implements the Middleware interface, just recover from the panic. Would provide information on the web page
// if in debug mode.
func (m *RecoveryWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, m.stackSize)
			stack = stack[:runtime.Stack(stack, m.stackAll)]
			log.Errorf("PANIC: %s\n%s", err, stack)
			if m.printStack {
				fmt.Fprintf(w, "PANIC: %s\n%s", err, stack)
			}
		}
	}()

	return next(ctx, w, r)
}

// NewRecoveryWare returns a new recovery middleware. Would log the full stack if enable the printStack.
func NewRecoveryWare(flags ...bool) Middleware {
	stackFlags := []bool{false, false}
	for i := range flags {
		if i >= len(stackFlags) {
			break
		}
		stackFlags[i] = flags[i]
	}
	return &RecoveryWare{
		printStack: stackFlags[0],
		stackAll:   stackFlags[1],
		stackSize:  1024 * 8,
	}
}

// StatWare is the statistics middleware which would log all the access and performation information.
type StatWare struct {
	ignoredPrefixes []string
}

// ServeHTTP implements the Middleware interface. Would log all the access, status and performance information.
func (m *StatWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
	start := time.Now()
	newCtx := next(ctx, w, r)
	res := w.(ResponseWriter)
	urlPath := r.URL.Path
	if res.Status() >= 400 {
		log.Warnf("Request %q %q, status=%v, size=%d, duration=%v",
			r.Method, r.URL.Path, res.Status(), res.Size(), time.Since(start))
	} else {
		ignored := false
		for _, prefix := range m.ignoredPrefixes {
			if strings.HasPrefix(urlPath, prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			log.Infof("Request %q %q, status=%v, size=%d, duration=%v",
				r.Method, r.URL.Path, res.Status(), res.Size(), time.Since(start))
		}
	}
	return newCtx
}

// NewStatWare returns a new StatWare, some ignored urls can be specified with prefixes which would not be logged.
func NewStatWare(prefixes ...string) Middleware {
	return &StatWare{prefixes}
}

type _OnionLayer struct {
	handler Middleware
	next    *_OnionLayer
}

func (m _OnionLayer) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return m.handler.ServeHTTP(ctx, w, r, m.next.ServeHTTP)
}

func buildOnion(wares []Middleware) _OnionLayer {
	var next _OnionLayer
	if len(wares) == 0 {
		return hollowOnion()
	} else if len(wares) > 1 {
		next = buildOnion(wares[1:])
	} else {
		next = hollowOnion()
	}
	return _OnionLayer{wares[0], &next}
}

func hollowOnion() _OnionLayer {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
		return ctx
	}
	return _OnionLayer{MiddleFn(fn), &_OnionLayer{}}
}
