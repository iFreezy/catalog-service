package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Repository Repository
}

type Repository struct {
	Postgres RepositoryPostgres
}

type RepositoryPostgres struct {
	Address        string        `envconfig:"APP_REPOSITORY_POSTGRES_ADDRESS"             default:"localhost:5429"`
	Username       string        `envconfig:"APP_REPOSITORY_POSTGRES_USERNAME"`
	Password       string        `envconfig:"APP_REPOSITORY_POSTGRES_PASSWORD"`
	Name           string        `envconfig:"APP_REPOSITORY_POSTGRES_NAME"`
	MigrationTable string        `envconfig:"APP_REPOSITORY_POSTGRES_MIGRATION_TABLE"     default:"schema_migrations"`
	ReadTimeout    time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_READ_TIMEOUT"        default:"30s"`
	WriteTimeout   time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_WRITE_TIMEOUT"       default:"30s"`
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
