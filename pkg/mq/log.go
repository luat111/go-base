package mq

import (
	"errors"
	"time"
)

const (
	PublishAction = "Publish"
	ConsumeAction = "Consume"
)

var (
	errClientConnIsNil = errors.New("RabbitMQ client connection is nil")
	errChannelIsNil    = errors.New("RabbitMQ channel connection is nil")
	errConsumerIsNil   = errors.New("RabbitMQ consumer connection is nil")
)

type LogMQMsg struct {
	CorrelationId string    `json:"correlationId"`
	Time          time.Time `json:"time"`
	Data          any       `json:"data,omitempty"`
	Error         error     `json:"error,omitempty"`
	Action        string    `json:"action,omitempty"`
}

func formatError(id, action string, time time.Time, data any, err error) LogMQMsg {
	return LogMQMsg{
		CorrelationId: id,
		Time:          time,
		Data:          data,
		Error:         err,
		Action:        action,
	}
}
