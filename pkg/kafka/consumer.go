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
	Key          string
}

func NewConsumer(conns *multiConn, conf config.Config, logger logger.ILogger, autoAck bool) (*Consumer, error) {
	readerCnf := getReaderConfig(conf)
	reader := newKafkaReader(readerCnf, conns.dialer)

	return &Consumer{
		ReaderConfig: readerCnf,
		Reader:       reader,
		Logger:       logger,
		AckOnConsume: autoAck,
		Key:          readerCnf.Key,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, fn HandlerFunc) {
	for {
		msg, err := c.Reader.FetchMessage(ctx)

		if err != nil {
			c.Logger.Error("Error when read", "err", err.Error())
		} else {
			attributes := headerToMap(msg.Headers)

			var data any
			json.Unmarshal(msg.Value, &data)
			logMsg := formatError(
				attributes[string(tracing.DefaultHeaderName)],
				ConsumeAction,
				time.Now(),
				data,
				nil,
			)

			if len(c.Key) > 0 && msg.Key != nil {
				ctx = context.WithValue(ctx, c.Key, string(msg.Key))
			}

			if c.AckOnConsume {
				c.Reader.CommitMessages(ctx, msg)
			}

			c.Logger.Info("Receive message", "Message", logMsg)

			fn(ctx, msg.Value, attributes)
		}
	}
}
