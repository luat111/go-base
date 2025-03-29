package grpc

import (
	"cmp"
	"context"
	"errors"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errUnavailable = errors.New("dial connection unavailable")

const roundRobinConfig = `{"loadBalancingConfig": [{"round_robin":{}}]}`

func CallRPC(logger logger.ILogger, ctx context.Context, method string, handler func(ctx context.Context) (any, error)) (any, error) {
	correlationId := cmp.Or(tracing.FromContext(ctx), tracing.DefaultGenerator())

	md := metadata.Pairs(GrpcTrackingId, correlationId)
	ctx = metadata.NewOutgoingContext(ctx, md)

	startTime := time.Now()
	res, err := handler(ctx)

	logRPC(Request, logger, correlationId, startTime, err, method, res)

	if err != nil {
		status, ok := status.FromError(err)

		if ok && status.Code() == codes.Unavailable {
			err = errUnavailable
		}
	}

	return res, err
}

func (g *GrpcServer) RegisterClient(name, addr string) error {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithDefaultServiceConfig(roundRobinConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		g.logger.Error(err)
		return err
	}

	g.logger.Info("Connected to RPC client via", "address", addr)

	g.Services[name] = conn
	return nil
}
