package respondent

import (
	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
)

type SimpleApplicator struct{}

func NewSimpleApplicator() *SimpleApplicator {
	return &SimpleApplicator{}
}

func (*SimpleApplicator) Apply(ctx any, manifest *Manifest) {
	type ManifestJSON struct {
		Status       int      `json:"-"`
		Error        string   `json:"error"`
		ErrorID      string   `json:"error_id,omitempty"`
		ErrorCode    int      `json:"error_code"`
		ErrorDetail  string   `json:"error_detail,omitempty"`
		ErrorDetails []string `json:"error_details,omitempty"`
	}

	if manifest == nil {
		return
	}

	httpCtx, ok := ctx.(HttpContext)
	if !ok {
		return
	}

	w := httpCtx.W
	r := httpCtx.R

	jsonManifest := (*ManifestJSON)(manifest)
	httph.SendEncoded(w, r, jsonManifest.Status, jsonManifest)
}

var _ Applicator = (*SimpleApplicator)(nil)
