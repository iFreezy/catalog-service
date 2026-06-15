package respondent

import (
	"errors"
	"net/http"
)

type SimpleExpander struct {
	rules    []extractorRule
	fallback ManifestExtractor
}

type extractorRule struct {
	pattern   error
	extractor ManifestExtractor
}

func NewSimpleExpander() *SimpleExpander {
	return &SimpleExpander{
		rules: make([]extractorRule, 0),
		fallback: func(err error) *Manifest {
			return &Manifest{
				Status:      http.StatusInternalServerError,
				Error:       "Internal server error",
				ErrorCode:   50001,
				ErrorDetail: err.Error(),
			}
		},
	}
}

func (se *SimpleExpander) Expand(err error) *Manifest {
	if err == nil {
		return nil
	}
	for _, rule := range se.rules {
		if errors.Is(err, rule.pattern) {
			if m := rule.extractor(err); m != nil {
				return m
			}
		}
	}
	if se.fallback == nil {
		return nil
	}
	return se.fallback(err)
}

func (se *SimpleExpander) ExtractorFor(sentinelErr error, extractor ManifestExtractor) *SimpleExpander {
	if sentinelErr == nil || extractor == nil {
		return se
	}
	se.rules = append(se.rules, extractorRule{pattern: sentinelErr, extractor: extractor})
	return se
}

func (se *SimpleExpander) FallbackExtractor(extractor ManifestExtractor) *SimpleExpander {
	se.fallback = extractor
	return se
}

func (se *SimpleExpander) WithoutDetail(err error, status, errorCode int, message string) *SimpleExpander {
	return se.ExtractorFor(err, func(_ error) *Manifest {
		return &Manifest{
			Status:    status,
			Error:     message,
			ErrorCode: errorCode,
		}
	})
}

func (se *SimpleExpander) WithDetail(err error, status, errorCode int, message, detail string) *SimpleExpander {
	return se.ExtractorFor(err, func(_ error) *Manifest {
		return &Manifest{
			Status:      status,
			Error:       message,
			ErrorCode:   errorCode,
			ErrorDetail: detail,
		}
	})
}
