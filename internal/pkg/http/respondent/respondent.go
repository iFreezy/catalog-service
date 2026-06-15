package respondent

import (
	"errors"
	"net/http"

	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
)

type respondent struct {
	expander   Expander
	replacer   Replacer
	applicator Applicator
}

type HttpContext struct {
	W http.ResponseWriter
	R *http.Request
}

var ErrBadExpander = errors.New("Respondent: expander is required")

func newRespondent(expander Expander, replacer Replacer, applicator Applicator) *respondent {
	if expander == nil {
		panic(ErrBadExpander)
	}
	if replacer == nil {
		replacer = NewSimpleReplacer()
	}
	if applicator == nil {
		applicator = NewSimpleApplicator()
	}
	return &respondent{
		expander:   expander,
		replacer:   replacer,
		applicator: applicator,
	}
}

func (rp *respondent) Callback(ctx any, err error) {
	err = rp.replacer.Replace(err)
	if err == nil {
		return
	}
	manifest := rp.expander.Expand(err)
	if manifest == nil {
		return
	}
	rp.applicator.Apply(ctx, manifest)
}

func (rp *respondent) CallbackForHTTP(w http.ResponseWriter, r *http.Request, err error) {
	rp.Callback(HttpContext{W: w, R: r}, err)
}

func NewMiddleware(expander Expander, replacer Replacer, applicator Applicator) httph.Middleware {
	re := newRespondent(expander, replacer, applicator)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			err := httph.ErrorGet(r)
			mayHandle := httph.ErrorTryAcquireHandling(r)
			if err != nil && mayHandle {
				re.CallbackForHTTP(w, r, err)
			}
		})
	}
}
