package utils

import (
	"context"
	"log"
	"time"
)

const TimeOutDuration = time.Second * 10
const TimeOutRetry = time.Second * 2

type ShutDownFunc func(context.Context) error

func GracefulShutDown(ctx context.Context, shutdownFn ShutDownFunc) error {
	err := shutdownFn(ctx)

	if err == context.DeadlineExceeded {
		log.Print("Halted active connections")
	}

	return err
}
