package kafka

import (
	"context"
	"go-base/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type kafkaMessage struct {
	msg    *kafka.Message
	reader Reader
	logger logger.ILogger
}

func newKafkaMessage(msg *kafka.Message, reader Reader, logger logger.ILogger) *kafkaMessage {
	return &kafkaMessage{
		msg:    msg,
		reader: reader,
		logger: logger,
	}
}

func (kmsg *kafkaMessage) Commit() {
	if kmsg.reader != nil {
		err := kmsg.reader.CommitMessages(context.Background(), *kmsg.msg)
		if err != nil {
			kmsg.logger.Error("unable to commit message on kafka")
		}
	}
}
