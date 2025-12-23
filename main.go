package main

import (
	"log"

	"github.com/iFreezy/catalog-service/internal/app/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	//log.Printf("Server will start on port: %d", cfg.Server.Port)
	log.Printf("Database DSN: postgresql://%s:$s@%s/%s",
		cfg.Repository.Postgres.Username,
		cfg.Repository.Postgres.Password,
		cfg.Repository.Postgres.Address,
		cfg.Repository.Postgres.Name)
}
