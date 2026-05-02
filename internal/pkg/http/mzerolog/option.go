package mzerolog

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Option = func(m *middleware)

func WithLogger(l zerolog.Logger) Option {
	return func(m *middleware) {
		m.log = l
	}
}

func WithSkipper(skipper func(r *http.Request) bool) Option {
	return func(m *middleware) {
		if skipper == nil {
			return
		}
		m.fromOptions.skipper = skipper
	}
}

func WithStringExtractor(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		ext := newStringExtractor(key, callback)
		m.fromOptions.extStrOnSuccess = append(m.fromOptions.extStrOnSuccess, ext)
		m.fromOptions.extStrOnFail = append(m.fromOptions.extStrOnFail, ext)
	}
}

func WithStringExtractorOnSuccess(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		m.fromOptions.extStrOnSuccess = append(m.fromOptions.extStrOnSuccess, newStringExtractor(key, callback))
	}
}

func WithStringExtractorOnFail(key string, callback CallbackExtractorString) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		m.fromOptions.extStrOnFail = append(m.fromOptions.extStrOnFail, newStringExtractor(key, callback))
	}
}

func WithAnyExtractor(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		ext := newAnyExtractor(key, callback)
		m.fromOptions.extAnyOnSuccess = append(m.fromOptions.extAnyOnSuccess, ext)
		m.fromOptions.extAnyOnFail = append(m.fromOptions.extAnyOnFail, ext)
	}
}

func WithAnyExtractorOnSuccess(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		m.fromOptions.extAnyOnSuccess = append(m.fromOptions.extAnyOnSuccess, newAnyExtractor(key, callback))
	}
}

func WithAnyExtractorOnFail(key string, callback CallbackExtractorAny) Option {
	return func(m *middleware) {
		if key == "" || callback == nil {
			return
		}
		m.fromOptions.extAnyOnFail = append(m.fromOptions.extAnyOnFail, newAnyExtractor(key, callback))
	}
}
