package mq

import (
	"context"
	"encoding/json"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "go-base/mq"

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
	// Start a producer span.
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "mq.publish "+msg.Route)
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", msg.ExchangeName),
		attribute.String("messaging.routing_key", msg.Route),
	)

	// Inject W3C trace context into the message headers so consumers can
	// extract and continue the trace.
	carrier := amqpHeaderCarrier(msg.Headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	// Also keep the existing correlation-id header for logging.
	cId := tracing.FromContext(ctx)
	msg.Headers[string(tracing.DefaultHeaderName)] = cId

	opts := MapToTable(msg.Headers)

	timestamp := time.Now()
	body, err := json.Marshal(msg.Body)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
