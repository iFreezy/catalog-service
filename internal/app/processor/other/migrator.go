package pprocessor

import (
	"context"
	"sync"

	"github.com/iFreezy/catalog-service/internal/app/processor"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	"github.com/rs/zerolog/log"
)

// procMigrate is a Processor that applies pending database migrations and then
// completes (unlike long-running processors such as the HTTP server).
type procMigrate struct {
	migrator repository.Migrate
}

func NewMigrator(migrator repository.Migrate) processor.Processor {
	return &procMigrate{migrator: migrator}
}

func (p *procMigrate) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.Wrap(ctx, wg, p.job)
}

func (p *procMigrate) job(ctx context.Context) {
	oldVer, newVer, err := p.migrator.Migrate(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to migrate repository schema")
		return
	}

	if oldVer != newVer {
		log.Info().
			Int64("old_ver", oldVer).
			Int64("new_ver", newVer).
			Msg("Repository schema has been updated")
		return
	}

	log.Info().Msg("Repository schema is up to date, nothing to migrate")
}
