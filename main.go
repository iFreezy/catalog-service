package main

import (
	"log"

	"github.com/iFreezy/catalog-service/internal/app/config"
	rhealth "github.com/iFreezy/catalog-service/internal/app/handler/health"
	rprocessor "github.com/iFreezy/catalog-service/internal/app/processor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Database DSN: postgresql://%s:%s@%s/%s",
		cfg.Repository.Postgres.Username,
		cfg.Repository.Postgres.Password,
		cfg.Repository.Postgres.Address,
		cfg.Repository.Postgres.Name)

	healthHandler := rhealth.NewHandler()

	httpServer := rprocessor.NewHttp(healthHandler, cfg.WebServer)

	if err := httpServer.Serve(); err != nil {
		log.Fatal(err)
	}
}
