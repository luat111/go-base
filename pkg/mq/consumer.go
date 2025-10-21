package mq

import (
	"context"
	"encoding/json"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

type HandlerFunc func(body []byte, metadata map[string]string, msg amqp091.Delivery)

type Consumer struct {
	Channel  *Channel
	Exchange *Exchange
	Queue    *Queue
	AutoAck  bool

	Logger logger.ILogger

	BoundRoute map[string]HandlerFunc
}

func newConsumer(
	client *RabbitClient,
	exchangeName, queueName string,
	autoAck bool,
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

	consumer := &Consumer{
		Channel:    channel,
		Exchange:   exchange,
		Queue:      queue,
		AutoAck:    autoAck,
		Logger:     client.Logger,
		BoundRoute: make(map[string]HandlerFunc),
	}

	client.Consumer = consumer

	return consumer
}

func (c *Consumer) BindQueue(route string, handler HandlerFunc) error {
	err := c.Channel.QueueBind(c.Queue.Name, route, c.Exchange.Name, false, nil)

	if err != nil {
		c.Logger.Error("Bind queue failed", "route", route, "err", err)
		return err
	}

	c.Logger.Info("Bind queue successfully", "route", route)

	c.BoundRoute[route] = handler

	return err
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
		handler := c.BoundRoute[msg.RoutingKey]

		if handler != nil {
			metadata := TableToMap(msg.Headers)

			var data any
			json.Unmarshal(msg.Body, &data)

			logMsg := formatError(
				metadata[string(tracing.DefaultHeaderName)],
				ConsumeAction,
				time.Now(),
				data,
				nil,
			)

			c.Logger.Info("Receive message", "Message", logMsg)

			handler(msg.Body, metadata, msg)

			if c.AutoAck {
				msg.Ack(false)
			}
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
