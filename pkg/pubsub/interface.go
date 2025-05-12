package pubsub

import "context"

type Writer interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

type Reader interface {
	Subscribe(ctx context.Context, topic string) (*Message, error)
}

type Client interface {
	Writer
	Reader

	CreateTopic(context context.Context, name string) error
	DeleteTopic(context context.Context, name string) error

	Close() error
}

type Committer interface {
	Commit()
}