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

func New(
	cfg section.WebServer,
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
) *Processor {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	vGenericRegHealthCheck(router, hHealth)

	rV1 := router.PathPrefix("/v1").Subrouter()

	if hCategory != nil {
		v1RegCategoryHandler(rV1, hCategory)
	}
	if hProduct != nil {
		v1RegProductHandler(rV1, hProduct)
	}

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
