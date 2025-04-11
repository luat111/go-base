package mq

import "go-base/pkg/logger"

const topicExchange = "topic"

type Exchange struct {
	Channel *Channel
	Name    string
}

func newExchange(channel *Channel, logger logger.ILogger, name string) (*Exchange, error) {
	err := channel.ExchangeDeclare(name, topicExchange, true, false, false, false, nil)

	if err != nil {
		logger.Error("Create exchange failed", "err", err)
		return nil, err
	}

	return &Exchange{
		Channel: channel,
		Name:    name,
	}, nil
}

func generateExchangeName(name string) string {
	return name + "_EXCHANGE"
}
