package mq

import "time"

const (
	PublishAction = "Publish"
	ConsumeAction = "Consume"
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
