package processor

import (
	"context"
	"io"
	"sync"
)

// Wrap runs cb in a separate goroutine, tracking it in wg (when provided) and
// skipping execution if ctx is already cancelled before the work starts.
func Wrap(ctx context.Context, wg *sync.WaitGroup, cb func(context.Context)) {
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		if wg != nil {
			defer wg.Done()
		}

		select {
		case <-ctx.Done():
			return
		default:
			cb(ctx)
		}
	}()
}

// WatchForShutdown waits for ctx to be cancelled and then releases the resource
// by calling closer.Close(). The goroutine is tracked in wg so that the caller
// can wait for every cleanup to finish before exiting.
func WatchForShutdown(ctx context.Context, wg *sync.WaitGroup, closer io.Closer) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		_ = closer.Close()
	}()
}
