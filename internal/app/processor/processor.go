package rprocessor

import (
	"github.com/iFreezy/catalog-service/internal/app/config/section"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
	processorhttp "github.com/iFreezy/catalog-service/internal/app/processor/http"
)

func NewHttp(health rhandler.Health, cfg section.WebServer) *processorhttp.Processor {
	return processorhttp.New(health, cfg)
}
