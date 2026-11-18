package util

import (
	"context"
	"time"
)

type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}

type CloserContextFunc = func(ctx context.Context) error

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
