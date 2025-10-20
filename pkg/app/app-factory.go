package app

import (
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/container"
	"go-base/pkg/grpc"
	"go-base/pkg/logger"
	"go-base/pkg/restful"
	"strconv"
)

func New[EnvInterface any](envOption config.EnvOptions) *App[EnvInterface] {
	app := &App[EnvInterface]{}
	app.LoadConfig(envOption)
	app.container = container.NewContainer(app.Config)
	app.logger = logger.NewLogger(common.AppPrefix)

	// HTTP Server
	port, err := strconv.Atoi(app.Config.Get(config.PORT))
	if err != nil || port <= 0 {
		port = common.DefaultHTTPPort
	}

	app.httpServer = restful.NewHTTPServer(app.container, app.Config, port, make(map[string]string))

	// GRPC Server
	rpcPort, err := strconv.Atoi(app.Config.Get(config.RPC_PORT))
	if err == nil || rpcPort > 0 {
		app.grpcRegistered = true
		app.grpcServer = grpc.NewGRPCServer(app.container, rpcPort)
	}

	return app
}
