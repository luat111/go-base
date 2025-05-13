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

type Consumer struct {
	ReaderConfig *ReaderConfig
	Reader       *kafka.Reader
	Logger       logger.ILogger
	AckOnConsume bool
	Apply        HandlerFunc
}

func NewConsumer(
	conns *multiConn,
	conf config.Config,
	logger logger.ILogger,
	topic string,
	handler HandlerFunc,
	autoAck bool,
) (*Consumer, error) {
	readerCnf := getReaderConfig(conf)
	reader := newKafkaReader(readerCnf, conns.dialer, topic)

	return &Consumer{
		ReaderConfig: readerCnf,
		Reader:       reader,
		Logger:       logger,
		AckOnConsume: autoAck,
		Apply:        handler,
	}, nil
}

func (c *Consumer) Consume() {
	for {
		ctx := context.Background()
		msg, err := c.Reader.FetchMessage(ctx)

		if err != nil {
			c.Logger.Error("Error when read", "err", err.Error())
		} else {
			attributes := headerToMap(msg.Headers)
			correlationId := attributes[string(tracing.DefaultHeaderName)]

			var data any
			json.Unmarshal(msg.Value, &data)
			logMsg := formatError(
				correlationId,
				ConsumeAction,
				time.Now(),
				data,
				nil,
			)

			if correlationId != "" {
				ctx = context.WithValue(ctx, tracing.DefaultHeaderName, correlationId)
			}

			if c.AckOnConsume {
				c.Reader.CommitMessages(ctx, msg)
			}

			c.Logger.Info("Receive message", "Message", logMsg)

			c.Apply(ctx, msg.Value, attributes)
		}
	}
}
