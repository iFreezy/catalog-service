package mzerolog

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
)

type CallbackExtractorString = func(r *http.Request) string

type CallbackExtractorAny = func(r *http.Request) any

type extractorStr struct {
	key string
	ext CallbackExtractorString
}

type extractorAny struct {
	key string
	ext CallbackExtractorAny
}

type middleware struct {
	log         zerolog.Logger
	fromOptions struct {
		extStrOnSuccess []extractorStr
		extAnyOnSuccess []extractorAny
		extStrOnFail    []extractorStr
		extAnyOnFail    []extractorAny
		skipper         func(r *http.Request) bool
	}
}

func (m *middleware) applyExtractors(
	r *http.Request,
	ev *zerolog.Event,
	extractorsString []extractorStr,
	extractorsAny []extractorAny,
) {
	for i, n := 0, len(extractorsString); i < n; i++ {
		e := extractorsString[i]
		if v := e.ext(r); v != "" {
			ev.Str(e.key, v)
		}
	}
	for i, n := 0, len(extractorsAny); i < n; i++ {
		e := extractorsAny[i]
		if v := e.ext(r); v != nil {
			ev.Any(e.key, v)
		}
	}
}

func (m *middleware) Callback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const (
			TailSuccess = " finished with no error"
			TailFail    = " finished (or aborted) with error"
		)

		start := time.Now()

		next.ServeHTTP(w, r)

		err := httph.ErrorGet(r)
		execTime := time.Since(start)

		if m.fromOptions.skipper(r) {
			return
		}

		var mb strings.Builder
		mb.Grow(48 + len(r.RequestURI))
		mb.WriteString(r.Method)
		mb.WriteByte(' ')
		mb.WriteString(r.RequestURI)

		var (
			ev        *zerolog.Event
			extString []extractorStr
			extAny    []extractorAny
		)
		if err == nil {
			mb.WriteString(TailSuccess)
			ev = m.log.Debug()
			extString = m.fromOptions.extStrOnSuccess
			extAny = m.fromOptions.extAnyOnSuccess
		} else {
			mb.WriteString(TailFail)
			ev = m.log.Error()
			extString = m.fromOptions.extStrOnFail
			extAny = m.fromOptions.extAnyOnFail
		}

		m.applyExtractors(r, ev, extString, extAny)

		ev.Err(err)
		ev.Str("exec_time", execTime.String())
		ev.Str("client_ip", r.RemoteAddr)

		ev.Msg(mb.String())
	})
}

func NewMiddleware(opts ...Option) httph.Middleware {
	m := middleware{
		log: log.Logger,
	}
	m.fromOptions.skipper = defaultSkipper

	for _, opt := range opts {
		opt(&m)
	}

	return m.Callback
}

func defaultSkipper(_ *http.Request) bool {
	return false
}

func newStringExtractor(key string, cb CallbackExtractorString) extractorStr {
	return extractorStr{key: key, ext: cb}
}

func newAnyExtractor(key string, cb CallbackExtractorAny) extractorAny {
	return extractorAny{key: key, ext: cb}
}
