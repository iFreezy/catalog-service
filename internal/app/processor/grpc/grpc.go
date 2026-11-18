package pgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/iFreezy/catalog-service/internal/app/processor"
	"github.com/iFreezy/catalog-service/internal/app/util"
	catalogv1 "github.com/iFreezy/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcProc struct {
	server *grpc.Server
	addr   string
}

func NewGRPC(catalogV1 catalogv1.CatalogServiceServer, cfg section.ProcessorGrpc) processor.Processor {
	server := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(server, catalogV1)
	reflection.Register(server)
	return &grpcProc{server: server, addr: fmt.Sprintf(":%d", cfg.ListenPort)}
}

func (p *grpcProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Str("listen_addr", p.addr).Msg("Failed to start gRPC server")
		return
	}
	log.Info().Str("listen_addr", p.addr).Msg("gRPC server listening")
	go func() { _ = p.server.Serve(listener) }()
	processor.WatchForShutdown(ctx, wg, util.CloserFunc(func() error {
		p.server.GracefulStop()
		return nil
	}))
}
