package mq

import (
	"context"
	"encoding/json"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Channel *Channel
	Logger  logger.ILogger
}

func newProducer(client *RabbitClient) *Producer {
	channel, _ := newChannel(client)
	producer := &Producer{Channel: channel, Logger: client.Logger}

	client.Producer = producer

	return producer
}

func (p *Producer) Publish(ctx context.Context, msg *Message) error {
	cId := tracing.FromContext(ctx)
	msg.Headers[string(tracing.DefaultHeaderName)] = cId

	opts := MapToTable(msg.Headers)

	timestamp := time.Now()
	body, err := json.Marshal(msg.Body)

	if err != nil {
		p.Logger.Error("Invalid data", "err", err)
		return err
	}

	err = p.Channel.Publish(
		msg.ExchangeName, // exchange name
		msg.Route,        // routing key
		false,            // mandatory
		false,            // immediate
		amqp091.Publishing{
			Headers:      opts,
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    timestamp,
			Body:         body,
		},
	)

	logMsg := formatError(cId, PublishAction, timestamp, msg.Body, err)

	if err != nil {
		p.Logger.Error("Publish message failed", "Message", logMsg)
	} else {
		p.Logger.Info("Message published", "Message", logMsg)
	}

	return err
}

func MapToTable(attributes map[string]string) amqp091.Table {
	opts := amqp091.Table{}

	for k, v := range attributes {
		opts[k] = v
	}

	return opts
}
