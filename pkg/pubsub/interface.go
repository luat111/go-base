package pubsub

import "context"

type Producer interface {
	Publish(ctx context.Context, msg Message) error
}

type Consumer interface {
	Consume(ctx context.Context) (*Message, error)
}

type Client interface {
	Producer
	Consumer

	CreateTopic(context context.Context, name string) error
	DeleteTopic(context context.Context, name string) error

	Close() error
}

type Message struct {
	ExchangeName string
	Route        string
	Body         any
	Headers      map[string]string
}
