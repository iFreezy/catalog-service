package builder

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/iFreezy/catalog-service/internal/app/config"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
	rhealth "github.com/iFreezy/catalog-service/internal/app/handler/health"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	pprocessor "github.com/iFreezy/catalog-service/internal/app/processor/other"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
	pcategory "github.com/iFreezy/catalog-service/internal/app/repository/postgres/category"
	pproduct "github.com/iFreezy/catalog-service/internal/app/repository/postgres/product"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Builder encapsulates the whole application-assembly logic: it holds every
// component and exposes Build* methods that initialize them in the right order,
// accumulating the first initialization error in b.err.
type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error

	cfg          config.Config
	connPostgres *rcpostgres.Client
	categoryRepo repository.Category
	productRepo  repository.Product

	healthHandler rhandler.Health

	processors []processor.Processor
}

func NewBuilder(cCtx *cli.Context) *Builder {
	var b = Builder{
		cCtx: cCtx,
		ctx:  context.Background(),
	}

	// Health Handler has no dependencies, so it is created right away.
	b.healthHandler = rhealth.NewHandler()

	return &b
}

func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	b.exec(true, func(b *Builder) {
		b.buildConfig(config.LoadArgs{}, injectors)
	})
}

func (b *Builder) BuildConfigSimple(injectors ...func(c *config.Config)) {
	b.exec(true, func(b *Builder) {
		b.buildConfig(config.LoadArgs{SkipConfig: true}, injectors)
	})
}

// Run starts all prepared processor.Processor and waits until they all
// are completed or an interrupt signal is received from the OS.
func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
	}

	log.Info().Msg("Application is initialized")
	defer log.Info().Msg("Application is completed, GoodBye!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(true, func(b *Builder) {
		conn, err := rcpostgres.NewConn(b.ctx, b.cfg.Repository.Postgres)
		if err != nil {
			b.err = fmt.Errorf("build postgres connection: %w", err)
			return
		}

		b.connPostgres = conn
	})
}

func (b *Builder) BuildRepoConnMigrator() {
	b.exec(b.connPostgres != nil, func(b *Builder) {
		proc := pprocessor.NewMigrator(b.connPostgres)
		b.processors = append(b.processors, proc)
	})
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORIES /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryRepo = pcategory.NewRepoFromPostgres(b.connPostgres)
	}, b.connPostgres)
}

func (b *Builder) BuildRepoProduct() {
	b.exec(true, func(b *Builder) {
		b.productRepo = pproduct.NewRepoFromPostgres(b.connPostgres)
	}, b.connPostgres)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE METHODS //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) buildConfig(args config.LoadArgs, injectors []func(*config.Config)) {
	args.Output = b.cCtx.App.Writer
	args.EnableSimpleLog = b.cCtx.Bool("no-json")

	config.Load(args)

	for i := range injectors {
		if injectors[i] != nil {
			injectors[i](&config.Root)
		}
	}

	b.cfg = config.Root
}

// exec runs cb only when preCond holds, no previous error occurred, and every
// dependency in requiredArgs is present (non-nil / non-zero). Otherwise it
// records a descriptive error in b.err so that subsequent Build* calls and Run
// short-circuit instead of operating on missing dependencies.
func (b *Builder) exec(preCond bool, cb func(b *Builder), requiredArgs ...any) {
	if !preCond || b.err != nil {
		return
	}

	for i, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependencies)", i)
			return
		}

		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}

		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}
