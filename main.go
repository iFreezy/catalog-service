package main

import (
	"context"
	"log"

	"github.com/iFreezy/catalog-service/internal/app/config"
	hcategory "github.com/iFreezy/catalog-service/internal/app/handler/category"
	rhealth "github.com/iFreezy/catalog-service/internal/app/handler/health"
	hproduct "github.com/iFreezy/catalog-service/internal/app/handler/product"
	rprocessor "github.com/iFreezy/catalog-service/internal/app/processor/http"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
	pcategory "github.com/iFreezy/catalog-service/internal/app/repository/postgres/category"
	pproduct "github.com/iFreezy/catalog-service/internal/app/repository/postgres/product"
	scategory "github.com/iFreezy/catalog-service/internal/app/service/category"
	sproduct "github.com/iFreezy/catalog-service/internal/app/service/product"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключение к PostgreSQL
	pgClient, err := rcpostgres.NewConn(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	log.Println("Successfully connected to PostgreSQL!")

	// Миграции
	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	if oldVer != newVer {
		log.Printf("Database migrated: old_version=%d, new_version=%d", oldVer, newVer)
	} else {
		log.Printf("Database is up to date: version=%d", newVer)
	}

	// Репозитории
	categoryRepo := pcategory.NewRepoFromPostgres(pgClient)
	productRepo := pproduct.NewRepoFromPostgres(pgClient)

	// Сервисы
	categorySvc := scategory.NewService(categoryRepo, productRepo)
	productSvc := sproduct.NewService(productRepo, categoryRepo)

	// Хендлеры
	healthHandler := rhealth.NewHandler()
	categoryHandler := hcategory.NewHandler(categorySvc)
	productHandler := hproduct.NewHandler(productSvc)

	// HTTP-сервер
	server := rprocessor.New(
		cfg.WebServer,
		healthHandler,
		categoryHandler,
		productHandler,
	)

	log.Printf("Starting HTTP server on %s", cfg.WebServer.Address)
	if err := server.Serve(); err != nil {
		log.Fatal("HTTP server error:", err)
	}
}
