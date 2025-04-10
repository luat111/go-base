package rabbit

import (
	"context"
	"go-base/pkg/logger"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

type Consumer struct {
	Channel      *Channel
	Exchange     *Exchange
	Queue        *Queue
	AutoAck      bool
	AckOnConsume bool
	Logger       logger.ILogger
}

func NewConsumer(
	client *RabbitClient,
	exchangeName, queueName string,
	autoAck, ackOnConsume bool,
) *Consumer {
	channel, _ := newChannel(client)

	var exchange *Exchange
	var queue *Queue

	g := new(errgroup.Group)

	g.Go(func() error {
		var err error
		exchange, err = newExchange(channel, client.Logger, exchangeName)
		return err
	})

	g.Go(func() error {
		var err error
		queue, err = newQueue(channel, client.Logger, queueName)
		return err
	})

	if err := g.Wait(); err != nil {
		client.Logger.Error("Init consumer failed", "err", err)
		return nil
	}

	return &Consumer{
		Channel:      channel,
		Exchange:     exchange,
		Queue:        queue,
		AutoAck:      autoAck,
		AckOnConsume: ackOnConsume,
		Logger:       client.Logger,
	}

}

func (c *Consumer) Consume() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgs, err := c.Channel.Consume(
		c.Queue.Name, // queue
		"",           // consumer
		c.AutoAck,    // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          //args
	)

	if err != nil {
		c.Logger.Error("Failed to consume", "err", err)
	}

	go c.ConsumeData(ctx, msgs)
}

func (c *Consumer) ConsumeData(ctx context.Context, messages <-chan amqp091.Delivery) {
	for msg := range messages {
		handler := c.Queue.BoundRoute[msg.RoutingKey]

		if handler != nil {
			if c.AckOnConsume && !c.AutoAck {
				msg.Ack(false)
			}

			metadata := TableToMap(msg.Headers)

			handler(msg.Body, metadata)
		}
	}
}

func TableToMap(header amqp091.Table) map[string]string {
	attributes := make(map[string]string, 0)
	for k, v := range header {
		attributes[k] = v.(string)
	}
	return attributes
}
