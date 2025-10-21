package mq

import (
	"github.com/rabbitmq/amqp091-go"
)

type Channel struct {
	*amqp091.Channel

	client *RabbitClient
}

func newChannel(client *RabbitClient) (*Channel, error) {
	if client == nil || client.conn == nil {
		return nil, errClientConnIsNil
	}

	channel, err := client.conn.Channel()

	if err != nil {
		client.Logger.Error("Create channel failed", "err", err)
		return nil, err
	}

	ch := &Channel{Channel: channel, client: client}

	return ch, nil
}
