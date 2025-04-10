package rabbit

import (
	"go-base/pkg/config"
	"go-base/pkg/logger"
)

type RabbitClient struct {
	conn   *RabbitConnection
	Logger logger.ILogger
}

func New(config config.Config, logger logger.ILogger) *RabbitClient {
	conn := newRabbitConn(config, logger)

	go conn.monitorConnection(logger)

	client := &RabbitClient{
		conn:   conn,
		Logger: logger,
	}

	return client
}
