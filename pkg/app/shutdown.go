package app

import (
	"context"
	"errors"
)

// Shutdown stops the service(s) and close the application.
// It shuts down the HTTP, gRPC, Metrics servers and closes the container's active connections to datasources.
func (a *App[EnvInterface]) Shutdown(ctx context.Context) error {
	var err error

	a.StopCron()

	// Flush and close the OTLP TracerProvider so in-flight spans are exported.
	if a.tracerShutdown != nil {
		a.tracerShutdown()
	}

	if a.httpServer != nil {
		err = errors.Join(err, a.httpServer.Shutdown(ctx))
	}

	if a.grpcServer != nil {
		err = errors.Join(err, a.grpcServer.Shutdown(ctx))
	}

	if a.container != nil {
		err = errors.Join(err, a.container.Close())
	}

	if err != nil {
		return err
	}

	a.logger.Info("Application shutdown complete")

	return err
}
