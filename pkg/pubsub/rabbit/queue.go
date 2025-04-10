package rabbit

import (
	"go-base/pkg/logger"

	"github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(body []byte, metadata map[string]string)

type Queue struct {
	*amqp091.Queue
	Channel    *Channel
	BoundRoute map[string]HandlerFunc
}

func newQueue(channel *Channel, logger logger.ILogger, name string) (*Queue, error) {
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

func (q *Queue) BindQueue(exchange *Exchange, route string, handler HandlerFunc) error {
	if exchange == nil {
		return nil
	}

	return nil
}
