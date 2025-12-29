package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/iFreezy/catalog-service/internal/app/config/section"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
)

type Processor struct {
	server *http.Server
}

func New(health rhandler.Health, cfg section.WebServer) *Processor {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	vGenericRegHealthCheck(router, health)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Processor{server: server}
}

func (p *Processor) Serve() error {
	return p.server.ListenAndServe()
}
