package kafka

import (
	"context"
	"encoding/json"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	conns *multiConn

	Logger   logger.ILogger
	Writer   *kafka.Writer
	Generate func() string
}

func NewProducer(conns *multiConn, conf config.Config, logger logger.ILogger, options ...func() string) (*Producer, error) {
	writerCnf := getWriterConfig(conf)
	writer := newKafkaWriter(writerCnf, conns.dialer)

	var generate func() string
	if len(options) > 0 {
		generate = options[0]
	}
	return &Producer{Writer: writer, Generate: generate, Logger: logger}, nil
}

func (p *Producer) Publish(ctx context.Context, data []byte, attributes map[string]string) error {
	var err error
	msg := kafka.Message{Value: data}

	cId := tracing.FromContext(ctx)
	timestamp := time.Now()
	body, err := json.Marshal(data)

	if attributes == nil {
		attributes = make(map[string]string)
		attributes[string(tracing.DefaultHeaderName)] = cId
	}

	msg.Headers = mapToHeader(attributes)

	if p.Generate != nil {
		id := p.Generate()
		msg.Key = []byte(id)
	}

	err = p.Writer.WriteMessages(ctx, msg)

	logMsg := formatError(cId, PublishAction, timestamp, body, err)

	if err != nil {
		p.Logger.Error("Publish message failed", "Message", logMsg)
	} else {
		p.Logger.Info("Message published", "Message", logMsg)
	}

	return err

}
