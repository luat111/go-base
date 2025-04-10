package pubsub

import "go-base/pkg/logger"

type Log struct {
	Mode          string `json:"mode"`
	CorrelationID string `json:"correlationID"`
	MessageValue  string `json:"messageValue"`
	Topic         string `json:"topic"`
	Host          string `json:"host"`
	PubSubBackend string `json:"pubSubBackend"`
	StartTime     string `json:"startTime"`
}

type PubsubLog struct {
	logger logger.ILogger
}

func New(log logger.ILogger) *PubsubLog {
	return &PubsubLog{
		logger: log,
	}
}

func (l *PubsubLog) Log(mode, correlationId, msg, topic, startTime string) {
	logMsg := Log{
		Mode:          mode,
		CorrelationID: correlationId,
		MessageValue:  msg,
		Topic:         topic,
		StartTime:     startTime,
	}

	l.logger.Info(logMsg)
}
