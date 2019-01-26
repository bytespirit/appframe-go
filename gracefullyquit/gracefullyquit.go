// Author: lipixun
// Created Time : 2019-01-26 16:13:12
//
// File Name: gracefullyquit.go
// Description:
//

package gracefullyquit

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefullQuiter implements the gracefully quit logic
type GracefullQuiter struct {
	liveContext    context.Context
	cancelLiveFunc context.CancelFunc
	handlers       []QuitHandler
	sigs           chan os.Signal
}

// QuitHandler defines a custom handler which will be executed when quitting
type QuitHandler interface {
	OnQuit()
}

// NewGracefullQuiter creates a new GracefullyQuiter
func NewGracefullQuiter(ctx context.Context, handlers ...QuitHandler) *GracefullQuiter {
	liveContext, cancelLiveFunc := context.WithCancel(ctx)
	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	quiter := &GracefullQuiter{liveContext, cancelLiveFunc, handlers, sigs}

	go func() {
		for {
			<-sigs
			quiter.StartQuit()
		}
	}()

	return quiter
}

// LiveContext returns the context that will be cancelled when quit signal received
func (q *GracefullQuiter) LiveContext() context.Context {
	return q.liveContext
}

// StartQuit starts a quit process immediately
func (q *GracefullQuiter) StartQuit() {
	q.cancelLiveFunc()
}

// WaitUntilExit waits until the application should exit immediately
// Args:
//	timeout 		The timeout of this wait method. 0 means no timeout
// Returns:
//	True means should exit
func (q *GracefullQuiter) WaitUntilExit(timeout time.Duration) bool {
	if timeout > 0 {
		select {
		case <-q.liveContext.Done():
			q.doExit()
			return true
		case <-time.After(timeout):
			return false
		}
	}
	<-q.liveContext.Done()
	q.doExit()
	return true
}

func (q *GracefullQuiter) doExit() {
	// Reset signals and close signal waiting thread
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM) // NOTE: NOT signal.reset(...)
	close(q.sigs)
	// Call all quit handlers
	if len(q.handlers) > 0 {
		for _, handler := range q.handlers {
			if handler != nil {
				handler.OnQuit()
			}
		}
	}
}

// QuitHandlerFunc defines the function of quit handler
type QuitHandlerFunc func()

type quitHandlerByFunc struct {
	f QuitHandlerFunc
}

func (h quitHandlerByFunc) OnQuit() {
	h.f()
}

// WithQuitHandlerFunc creates a QuitHandler by function
func WithQuitHandlerFunc(f QuitHandlerFunc) QuitHandler {
	return quitHandlerByFunc{f}
}
