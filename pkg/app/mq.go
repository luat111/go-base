package app

import (
	"go-base/pkg/mq"
)

func (a *App[EnvInterface]) ListenRMQ(handlers map[string]mq.HandlerFunc) error {
	return a.container.MQ.Listen(handlers)
}
