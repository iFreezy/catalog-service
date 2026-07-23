package util

import (
	"context"
	"time"
)

// CloserFunc is an adapter that lets an ordinary func() error satisfy io.Closer.
type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}

// CloserContextFunc is a shutdown-like function that needs a context,
// e.g. (*http.Server).Shutdown.
type CloserContextFunc = func(ctx context.Context) error

// NewCloserContextFunc adapts a context-aware function to CloserFunc, applying
// timeout to the call so that a graceful shutdown cannot hang forever.
// A non-positive timeout means "no deadline".
func NewCloserContextFunc(ctx context.Context, f CloserContextFunc, timeout time.Duration) CloserFunc {
	return func() error {
		callCtx := ctx

		if timeout > 0 {
			var cancelFunc context.CancelFunc

			callCtx, cancelFunc = context.WithTimeout(ctx, timeout)
			defer cancelFunc()
		}

		return f(callCtx)
	}
}
