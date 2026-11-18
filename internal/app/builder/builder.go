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
	ghcatalogv1 "github.com/iFreezy/catalog-service/internal/app/handler/grpc/catalog/v1"
	rhealth "github.com/iFreezy/catalog-service/internal/app/handler/health"
	hproduct "github.com/iFreezy/catalog-service/internal/app/handler/product"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	pgateway "github.com/iFreezy/catalog-service/internal/app/processor/gateway"
	pgrpc "github.com/iFreezy/catalog-service/internal/app/processor/grpc"
	rprocessor "github.com/iFreezy/catalog-service/internal/app/processor/http"
	pprocessor "github.com/iFreezy/catalog-service/internal/app/processor/other"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
	pcategory "github.com/iFreezy/catalog-service/internal/app/repository/postgres/category"
	pproduct "github.com/iFreezy/catalog-service/internal/app/repository/postgres/product"
	"github.com/iFreezy/catalog-service/internal/app/service"
	scategory "github.com/iFreezy/catalog-service/internal/app/service/category"
	sproduct "github.com/iFreezy/catalog-service/internal/app/service/product"
	catalogv1 "github.com/iFreezy/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const chErrorsBufferSize = 4096

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

	healthHandler    rhandler.Health
	categoryHandler  rhandler.Category
	productHandler   rhandler.Product
	catalogV1Handler catalogv1.CatalogServiceServer

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

func (b *Builder) BuildHandlerGrpcCatalogV1() {
	b.exec(true, func(b *Builder) {
		b.catalogV1Handler = ghcatalogv1.NewHandler(b.productService)
	}, b.productService)
}

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

func (b *Builder) BuildProcGrpc() {
	b.exec(true, func(b *Builder) {
		b.processors = append(b.processors, pgrpc.NewGRPC(b.catalogV1Handler, b.cfg.Processor.Grpc))
	}, b.catalogV1Handler)
}

func (b *Builder) BuildProcGateway() {
	b.exec(true, func(b *Builder) {
		b.processors = append(b.processors, pgateway.NewGateway(b.cfg.Processor.Gateway, b.cfg.Processor.Grpc))
	})
}

func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	var s = <-sig

	log.Info().Str("signal", s.String()).Msg("Shutdown is requested")

	cancelFunc()
}

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
