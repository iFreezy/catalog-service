package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Repository section.Repository
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		var pathError *os.PathError
		if !errors.As(err, &pathError) {
			return Config{}, fmt.Errorf("load .env: %w", err)
		}
	}

	var cfg Config

	if err := envconfig.Process("REPOSITORY", &cfg.Repository); err != nil {
		return Config{}, fmt.Errorf("parse REPOSITORY config: %w", err)
	}

	return cfg, nil
}
