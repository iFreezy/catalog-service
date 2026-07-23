package processor

import (
	"context"
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
