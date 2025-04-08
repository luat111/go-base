package app

import (
	"context"
	"go-base/pkg/common"
	"go-base/pkg/grpc"
	"go-base/pkg/restful"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func (a *App[EnvInterface]) Run() {
	// Start app
	go a.startServer()

	// Context for shutting down
	signChan, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-signChan.Done()

	// Create a shutdown context with a timeout
	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(signChan), common.DefaultTimeOut)
	defer done()

	a.container.Logger.Info("Shutting down server")
	if shutdownErr := a.Shutdown(shutdownCtx); shutdownErr != nil {
		a.container.Logger.Error("Server shutdown failed", "err", shutdownErr)
	}
}

func (a *App[EnvInterface]) startServer() {
	wg := sync.WaitGroup{}

	// Start HTTP Server
	if a.httpRegistered {
		wg.Add(1)
		a.httpServer.MappingRoutes()

		go func(s *restful.HttpServer) {
			defer wg.Done()
			s.Run(a.container, a.Config)
		}(a.httpServer)
	}

	// Start gRPC Server only if a service is registered
	if a.grpcRegistered {
		wg.Add(1)

		go func(s *grpc.GrpcServer) {
			defer wg.Done()
			s.Run(a.container)
		}(a.grpcServer)
	}

	wg.Wait()
}
