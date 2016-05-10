package server

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/getsentry/raven-go"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

// SentryRecoveryWare is the recovery middleware which can cover the panic situation.
// If a sentry client is in the context, it will send the panic to the sentry server.
type SentryRecoveryWare struct {
	printStack bool
	stackAll   bool
	stackSize  int
	client     *raven.Client
}

// ServeHTTP implements the Middleware interface, just recover from the panic.
// Would send information to the sentry server.
func (m *SentryRecoveryWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, m.stackSize)
			stack = stack[:runtime.Stack(stack, m.stackAll)]
			log.Errorf("PANIC: %s\n%s", err, stack)
			m.client.CaptureError(errors.New(fmt.Sprint(err)), nil)
			if m.printStack {
				fmt.Fprintf(w, "PANIC: %s\n%s", err, stack)
			}
		}
	}()

	return next(ctx, w, r)
}

// NewSentryRecoveryWare returns a new recovery middleware similar as RecoveryWare but can send messages to the sentry server.
func NewSentryRecoveryWare(client *raven.Client, flags ...bool) Middleware {
	stackFlags := []bool{false, false}
	for i := range flags {
		if i >= len(stackFlags) {
			break
		}
		stackFlags[i] = flags[i]
	}
	return &SentryRecoveryWare{
		client:     client,
		printStack: stackFlags[0],
		stackAll:   stackFlags[1],
		stackSize:  1024 * 8,
	}
}
