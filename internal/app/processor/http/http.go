package http

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/iFreezy/catalog-service/internal/app/config/section"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	"github.com/iFreezy/catalog-service/internal/app/util"
	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
	"github.com/iFreezy/catalog-service/internal/pkg/http/mzerolog"
	"github.com/rs/zerolog/log"
)

// shutdownTimeout bounds how long the server may take to finish in-flight
// requests once a shutdown has been requested.
const shutdownTimeout = 5 * time.Second

type httpProc struct {
	addr   string
	server *http.Server
}

func NewHttp(
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
	_ []httph.Middleware,
	cfg section.WebServer,
) processor.Processor {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	router.Use(httph.NewErrorMiddleware())

	router.Use(makeErrorMiddleware())

	router.Use(mzerolog.NewMiddleware(
		mzerolog.WithSkipper(util.IsFilteredHttpRoute),
		mzerolog.WithStringExtractor("user_id", func(r *http.Request) string {
			return r.Header.Get("X-User-ID")
		}),
		mzerolog.WithStringExtractor("session_id", func(r *http.Request) string {
			return r.Header.Get("X-Session-ID")
		}),
		mzerolog.WithStringExtractorOnFail("request_id", func(r *http.Request) string {
			return r.Header.Get("X-Request-ID")
		}),
		mzerolog.WithAnyExtractorOnSuccess("content_length", func(r *http.Request) any {
			if r.ContentLength > 0 {
				return r.ContentLength
			}
			return nil
		}),
	))

	vGenericRegHealthCheck(router, hHealth)

	rV1 := router.PathPrefix("/v1").Subrouter()

	if hCategory != nil {
		v1RegCategoryHandler(rV1, hCategory)
	}
	if hProduct != nil {
		v1RegProductHandler(rV1, hProduct)
	}

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &httpProc{addr: cfg.Address, server: server}
}

// StartAsync opens the TCP listener and serves it in background, registering
// the two-stage graceful shutdown: first the listener stops accepting new
// connections, then the server drains in-flight requests within shutdownTimeout.
func (p *httpProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig

	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Str("listen_addr", p.addr).
			Msg("Failed to start listening TCP addr for HTTP server")

		return
	}

	log.Info().Str("listen_addr", p.addr).
		Msg("Listening of TCP addr for HTTP server has been started")

	go p.serve(l)

	processor.WatchForShutdown(ctx, wg, util.CloserFunc(l.Close))
	processor.WatchForShutdown(ctx, wg, util.NewCloserContextFunc(
		context.Background(), p.server.Shutdown, shutdownTimeout,
	))
}

// serve runs the HTTP server and blocks the current goroutine until the
// listener is closed.
func (p *httpProc) serve(l net.Listener) {
	_ = p.server.Serve(l) // blocks the goroutine
}
