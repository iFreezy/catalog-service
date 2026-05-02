package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

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

	config.Load(config.LoadArgs{
		Output:          os.Stdout,
		EnableSimpleLog: true,
		SkipConfig:      false,
	})

	cfg := config.Root

	// Подключение к PostgreSQL
	pgClient, err := rcpostgres.NewConn(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to PostgreSQL")
	}
	log.Info().Msg("Successfully connected to PostgreSQL")

	// Миграции
	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}
	if oldVer != newVer {
		log.Info().Int64("old", oldVer).Int64("new", newVer).Msg("Database migrated")
	} else {
		log.Info().Int64("version", newVer).Msg("Database is up to date")
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

	log.Info().Str("addr", cfg.WebServer.Address).Msg("Starting HTTP server")
	if err := server.Serve(); err != nil {
		log.Fatal().Err(err).Msg("HTTP server error")
	}
}
