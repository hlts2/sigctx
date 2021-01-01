package sigctx

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
)

var (
	// TerminateLimit is a maximum count of received signals.
	// If the limit is reached, the process will be forcibly shutdown.
	TerminateLimit = 1024

	// Log is used for logging.
	Log Logger = new(log.Logger)
)

// Logger represents an interface for logging.
type Logger interface {
	Print(v ...interface{})
}

// WithCancel returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the parent context's Done channel is closed or an os signal is received, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithCancel(parent context.Context, sigs ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)

	var done int32

	go func() {
		defer close(sigCh)
		defer signal.Stop(sigCh)

		var retries int

		select {
		case <-parent.Done():
			return
		case <-ctx.Done():
			if atomic.LoadInt32(&done) > 1 {
				return
			}
		case sig := <-sigCh:
			retries++
			cancel()

			switch {
			case retries == 1:
				Log.Print(fmt.Sprintf("received %v, graceful shutdown", sig))
			case retries >= TerminateLimit:
				Log.Print(fmt.Sprintf("received %v, signals %d times, aborting", sig, retries))
				return
			default:
				Log.Print(fmt.Sprintf("received %v", sig))
			}
		}
	}()

	return ctx, func() {
		atomic.AddInt32(&done, 1)
		cancel()
	}
}
