package rabbit

import "go-base/pkg/logger"

const topicExchange = "topic"

type Exchange struct {
	Channel      *Channel
	ExchangeName string
}

func newExchange(channel *Channel, logger logger.ILogger, name string) (*Exchange, error) {
	err := channel.ExchangeDeclare(name, topicExchange, true, false, false, false, nil)

	if err != nil {
		logger.Error("Create exchange failed", "err", err)
		return nil, err
	}

	return &Exchange{
		Channel:      channel,
		ExchangeName: name,
	}, nil
}
