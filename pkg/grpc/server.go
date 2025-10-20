package grpc

import (
	"context"
	"go-base/pkg/common"
	"go-base/pkg/common/utils"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"net"
	"strconv"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcServer struct {
	Server    *grpc.Server
	port      int
	Services  map[string]*grpc.ClientConn
	container *container.Container
	logger    logger.ILogger
}

func NewGRPCServer(ctn *container.Container, port int) *GrpcServer {
	return &GrpcServer{
		Server: grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(
				keepalive.EnforcementPolicy{
					MinTime:             5 * time.Second,
					PermitWithoutStream: true,
				},
			),
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					grpc_recovery.UnaryServerInterceptor(),
					ObservabilityInterceptor(ctn.Logger),
				))),
		Services:  make(map[string]*grpc.ClientConn),
		port:      port,
		container: ctn,
		logger:    logger.NewLogger(common.RPCPrefix),
	}
}

func (g *GrpcServer) Run(c *container.Container) {
	address := ":" + strconv.Itoa(g.port)

	g.logger.Info("GRPC server is running at address", "address", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		g.logger.Error("error in starting gRPC server at", "address", address, "err", err)
		return
	}

	if err := g.Server.Serve(listener); err != nil {
		g.logger.Error("error in starting gRPC server at", "address", address, "err", err)
		return
	}
}

func (g *GrpcServer) Shutdown(ctx context.Context) error {
	return utils.GracefulShutDown(ctx, func(_ context.Context) error {
		g.Server.GracefulStop()

		return nil
	})
}
