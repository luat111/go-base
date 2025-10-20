package mq

import (
	"go-base/pkg/logger"

	"github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	*amqp091.Queue
	Channel *Channel
}

func newQueue(channel *Channel, logger logger.ILogger, name string) (*Queue, error) {
	if channel == nil {
		return nil, errChannelIsNil
	}
	
	queue, err := channel.QueueDeclare(name, true, false, false, false, nil)

	if err != nil {
		logger.Error("Create exchange failed", "err", err)
		return nil, err
	}

	return &Queue{
		Queue:   &queue,
		Channel: channel,
	}, nil
}

func generateQueueName(name string) string {
	return name + "_QUEUE"
}
