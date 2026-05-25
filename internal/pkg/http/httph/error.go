package httph

import (
	"context"
	"net/http"
)

type _ContextKeyError struct{}

type _ContextValueError struct {
	err       error
	detail    string
	isHandled bool
}

func errorPrepare(ctx context.Context) context.Context {
	var errCtx = new(_ContextValueError)
	return context.WithValue(ctx, _ContextKeyError{}, errCtx)
}

func errorGet(ctx context.Context) error {
	errV, ok := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if !ok || errV == nil {
		return nil
	}
	return errV.err
}

func errorGetDetail(ctx context.Context) string {
	errV, ok := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if !ok || errV == nil {
		return ""
	}
	return errV.detail
}

func errorApply(ctx context.Context, err error) {
	errV, ok := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if !ok || errV == nil {
		return
	}
	errV.err = err
}

func errorApplyDetail(ctx context.Context, detail string) {
	errV, ok := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if !ok || errV == nil {
		return
	}
	errV.detail = detail
}

func errorTryAcquireHandling(ctx context.Context) bool {
	errV, ok := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if !ok || errV == nil || errV.isHandled {
		return false
	}
	errV.isHandled = true
	return true
}

func ErrorPrepare(r *http.Request) *http.Request {
	return r.WithContext(errorPrepare(r.Context()))
}

func ErrorGet(r *http.Request) error {
	return errorGet(r.Context())
}

func ErrorGetDetail(r *http.Request) string {
	return errorGetDetail(r.Context())
}

func ErrorApply(r *http.Request, err error) {
	errorApply(r.Context(), err)
}

func ErrorApplyDetail(r *http.Request, detail string) {
	errorApplyDetail(r.Context(), detail)
}

func ErrorTryAcquireHandling(r *http.Request) bool {
	return errorTryAcquireHandling(r.Context())
}

type Middleware = func(http.Handler) http.Handler

func NewErrorMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, ErrorPrepare(r))
		})
	}
}
