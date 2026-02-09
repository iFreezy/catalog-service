package main

import (
	"context"
	"log"

	"github.com/iFreezy/catalog-service/internal/app/config"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Database DSN: postgresql://%s:%s@%s/%s",
		cfg.Repository.Postgres.Username,
		cfg.Repository.Postgres.Password,
		cfg.Repository.Postgres.Address,
		cfg.Repository.Postgres.Name)

	pgClient, err := rcpostgres.NewConn(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	log.Println("Successfully connected to PostgreSQL!")

	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	if oldVer != newVer {
		log.Printf("Database migrated: old_version=%d, new_version=%d", oldVer, newVer)
	} else {
		log.Printf("Database is up to date: version=%d", newVer)
	}
}
