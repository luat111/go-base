package rabbit

import (
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Channel *Channel
}

func NewProducer(client *RabbitClient) *Producer {
	channel, _ := newChannel(client)
	return &Producer{Channel: channel}
}

type Message struct {
	ExchangeName string
	Route        string
	Body         []byte
	Headers      map[string]string
}

func (p *Producer) PublishMessage(msg *Message) error {
	opts := MapToTable(msg.Headers)
	return p.Channel.Publish(
		msg.ExchangeName, // exchange name
		msg.Route,        // routing key
		false,            // mandatory
		false,            // immediate
		amqp091.Publishing{
			Headers:      opts,
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
			Body:         msg.Body,
		},
	)
}

func MapToTable(attributes map[string]string) amqp091.Table {
	opts := amqp091.Table{}

	for k, v := range attributes {
		opts[k] = v
	}

	return opts
}
