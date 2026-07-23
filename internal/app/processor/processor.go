package processor

import (
	"context"
	"sync"
)

// Processor is a component that performs some background work of the
// application (HTTP server, migrator, event consumer, cron job, etc.).
type Processor interface {
	StartAsync(ctx context.Context, wg *sync.WaitGroup)
}

// ProcessorFunc adapts an ordinary function to the Processor interface.
//
//goland:noinspection GoNameStartsWithPackageName
type ProcessorFunc func(ctx context.Context, wg *sync.WaitGroup)

func (p ProcessorFunc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	p(ctx, wg)
}
