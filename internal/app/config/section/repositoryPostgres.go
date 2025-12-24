package section

import "time"

type RepositoryPostgres struct {
	Address        string        `envconfig:"APP_REPOSITORY_POSTGRES_ADDRESS"             default:"localhost:5429"`
	Username       string        `envconfig:"APP_REPOSITORY_POSTGRES_USERNAME"`
	Password       string        `envconfig:"APP_REPOSITORY_POSTGRES_PASSWORD"`
	Name           string        `envconfig:"APP_REPOSITORY_POSTGRES_NAME"`
	MigrationTable string        `envconfig:"APP_REPOSITORY_POSTGRES_MIGRATION_TABLE"     default:"schema_migrations"`
	ReadTimeout    time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_READ_TIMEOUT"        default:"30s"`
	WriteTimeout   time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_WRITE_TIMEOUT"       default:"30s"`
}
