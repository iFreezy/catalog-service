package pgateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	"github.com/iFreezy/catalog-service/internal/app/util"
	catalogv1 "github.com/iFreezy/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	gatewayReadHeaderTimeout = 5 * time.Second
	gatewayReadTimeout       = 30 * time.Second
	gatewayWriteTimeout      = 30 * time.Second
	gatewayIdleTimeout       = 120 * time.Second
	gatewayShutdownTimeout   = 5 * time.Second
)

type gatewayProc struct {
	server   http.Server
	addr     string
	grpcAddr string
}

func NewGateway(cfgGateway section.ProcessorGateway, cfgGrpc section.ProcessorGrpc) processor.Processor {
	addr := fmt.Sprintf(":%d", cfgGateway.ListenPort)
	return &gatewayProc{
		server: http.Server{
			Addr:              addr,
			ReadHeaderTimeout: gatewayReadHeaderTimeout,
			ReadTimeout:       gatewayReadTimeout,
			WriteTimeout:      gatewayWriteTimeout,
			IdleTimeout:       gatewayIdleTimeout,
		},
		addr:     addr,
		grpcAddr: net.JoinHostPort("localhost", strconv.Itoa(int(cfgGrpc.ListenPort))),
	}
}

func (p *gatewayProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, p.grpcAddr, opts); err != nil {
		log.Fatal().Err(err).Msg("Failed to register gRPC Gateway routes")
		return
	}
	p.server.Handler = mux
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Str("listen_addr", p.addr).Msg("Failed to start gRPC Gateway")
		return
	}
	log.Info().Str("listen_addr", p.addr).Msg("gRPC Gateway listening")
	go func() { _ = p.server.Serve(listener) }()
	processor.WatchForShutdown(ctx, wg, util.CloserFunc(listener.Close))
	processor.WatchForShutdown(ctx, wg, util.NewCloserContextFunc(context.Background(), p.server.Shutdown, gatewayShutdownTimeout))
}
