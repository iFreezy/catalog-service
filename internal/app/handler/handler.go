package rhandler

import "net/http"

type (
	Health interface {
		Check(w http.ResponseWriter, r *http.Request)
	}
)
