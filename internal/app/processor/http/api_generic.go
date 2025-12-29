package http

import (
	"net/http"

	"github.com/gorilla/mux"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
)

func vGenericRegHealthCheck(r *mux.Router, h rhandler.Health) {
	reg(r, http.MethodGet, "/health", http.HandlerFunc(h.Check))
}

func handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
