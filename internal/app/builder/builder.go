package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/iFreezy/catalog-service/internal/app/config"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
	hcategory "github.com/iFreezy/catalog-service/internal/app/handler/category"
	rhealth "github.com/iFreezy/catalog-service/internal/app/handler/health"
	hproduct "github.com/iFreezy/catalog-service/internal/app/handler/product"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	rprocessor "github.com/iFreezy/catalog-service/internal/app/processor/http"
	pprocessor "github.com/iFreezy/catalog-service/internal/app/processor/other"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
	pcategory "github.com/iFreezy/catalog-service/internal/app/repository/postgres/category"
	pproduct "github.com/iFreezy/catalog-service/internal/app/repository/postgres/product"
	"github.com/iFreezy/catalog-service/internal/app/service"
	scategory "github.com/iFreezy/catalog-service/internal/app/service/category"
	sproduct "github.com/iFreezy/catalog-service/internal/app/service/product"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// chErrorsBufferSize lets asynchronous processors report errors without
// blocking when they produce them faster than they are logged.
const chErrorsBufferSize = 4096

// Builder encapsulates the whole application-assembly logic: it holds every
// component and exposes Build* methods that initialize them in the right order,
// accumulating the first initialization error in b.err.
type Builder struct {
	cCtx     *cli.Context
	ctx      context.Context
	wg       sync.WaitGroup
	err      error
	chErrors chan error

	cfg          config.Config
	connPostgres *rcpostgres.Client
	categoryRepo repository.Category
	productRepo  repository.Product

	categoryService service.Category
	productService  service.Product

	healthHandler   rhandler.Health
	categoryHandler rhandler.Category
	productHandler  rhandler.Product

	processors []processor.Processor
}

func NewBuilder(cCtx *cli.Context) *Builder {
	var b = Builder{
		cCtx:     cCtx,
		chErrors: make(chan error, chErrorsBufferSize),
	}

	var cancelFunc func()

	b.ctx, cancelFunc = context.WithCancel(context.Background())

	var sig = make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(sig, cancelFunc)
	go b.printErrors()

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
///// SERVICES /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildServiceCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryService = scategory.NewService(b.categoryRepo, b.productRepo)
	}, b.categoryRepo, b.productRepo)
}

func (b *Builder) BuildServiceProduct() {
	b.exec(true, func(b *Builder) {
		b.productService = sproduct.NewService(b.productRepo, b.categoryRepo)
	}, b.productRepo, b.categoryRepo)
}

////////////////////////////////////////////////////////////////////////////////
///// HANDLERS /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildHandlerHttpCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryHandler = hcategory.NewHandler(b.categoryService)
	}, b.categoryService)
}

func (b *Builder) BuildHandlerHttpProduct() {
	b.exec(true, func(b *Builder) {
		b.productHandler = hproduct.NewHandler(b.productService)
	}, b.productService)
}

////////////////////////////////////////////////////////////////////////////////
///// PROCESSORS ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		var procHttp = rprocessor.NewHttp(
			b.healthHandler,
			b.categoryHandler,
			b.productHandler,
			nil,
			b.cfg.WebServer,
		)

		b.processors = append(b.processors, procHttp)
	}, b.categoryHandler, b.productHandler)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE METHODS //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// waitForSignal blocks until the OS asks the application to stop and then
// cancels the application context, which triggers graceful shutdown of every
// running processor.
func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	var s = <-sig

	log.Info().Str("signal", s.String()).Msg("Shutdown is requested")

	cancelFunc()
}

// printErrors logs errors reported asynchronously by processors after start.
func (b *Builder) printErrors() {
	for err := range b.chErrors {
		log.Error().Err(err).Msg("Asynchronous error occurred")
	}
}

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
