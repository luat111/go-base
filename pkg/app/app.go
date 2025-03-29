package app

import (
	"go-base/pkg/config"
	"go-base/pkg/container"
	"go-base/pkg/grpc"
	"go-base/pkg/restful"
)

type App[EnvInterface any] struct {
	Config config.Config

	grpcServer *grpc.GrpcServer
	httpServer *restful.HttpServer
	// metricServer *metricServer

	// cron *Crontab

	// container is unexported because this is an internal implementation and applications are provided access to it via Context
	container *container.Container

	grpcRegistered bool
	httpRegistered bool

	// subscriptionManager SubscriptionManager
}

func (a *App[EnvInterface]) LoadConfig(opt config.EnvOptions) {
	config := config.NewAppConfig[EnvInterface](opt)

	a.Config = config
}
