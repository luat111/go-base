package app

import (
	"go-base/pkg/kafka"

	"golang.org/x/sync/errgroup"
)

type KafkaHandler struct {
	handler kafka.HandlerFunc
	autoAck bool
}

func (a *App[EnvInterface]) ListenKafka(handlers map[string]KafkaHandler) error {
	g := new(errgroup.Group)

	for topic, v := range handlers {
		g.Go(func() error {
			return a.container.Kafka.InitConsumer(a.Config, topic, v.handler, v.autoAck)
		})
	}

	if err := g.Wait(); err != nil {
		a.logger.Error("Init Kafka Consumers failed", "err", err)
		return err
	}

	a.container.Kafka.StartConsume()

	return nil
}
