package app

import (
	"go-base/pkg/mq"

	"golang.org/x/sync/errgroup"
)

func (a *App[EnvInterface]) ListenRMQ(handlers map[string]mq.HandlerFunc) error {
	g := new(errgroup.Group)

	for route, handler := range handlers {
		g.Go(func() error {
			return a.container.MQ.BindQueue(route, handler)
		})
	}

	if err := g.Wait(); err != nil {
		a.logger.Error("Listen RMQ failed", "err", err)
		return err
	}

	a.container.MQ.Consume()

	return nil
}
