package rhealth

import (
	"net/http"

	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
)

type HealthHandler struct{}

func NewHandler() rhandler.Health {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
