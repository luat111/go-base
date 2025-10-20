package app

import (
	"go-base/pkg/config"
	"go-base/pkg/container"
	"go-base/pkg/grpc"
	"go-base/pkg/logger"
	"go-base/pkg/restful"
)

type App[EnvInterface any] struct {
	Config config.Config

	grpcServer *grpc.GrpcServer
	httpServer *restful.HttpServer
	// metricServer *metricServer

	container *container.Container
	logger    logger.ILogger

	grpcRegistered bool
	httpRegistered bool
}

func (a *App[EnvInterface]) LoadConfig(opt config.EnvOptions) {
	config := config.NewAppConfig[EnvInterface](opt)

	a.Config = config
}
