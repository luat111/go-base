package kafka

import (
	"context"
	"go-base/pkg/logger"
	"go-base/pkg/pubsub"
	"go-base/pkg/tracing"
	"time"

	"github.com/segmentio/kafka-go"
)

func createKafkaWriter(conf *Config, dialer *kafka.Dialer, logger logger.ILogger) Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      conf.Brokers,
		Dialer:       dialer,
		BatchSize:    conf.BatchSize,
		BatchBytes:   conf.BatchBytes,
		BatchTimeout: time.Duration(conf.BatchTimeout),
		Logger:       kafka.LoggerFunc(logger.Debugf),
	})
}

func (k *KafkaClient) Publish(ctx context.Context, topic string, message []byte) error {
	trackingId := tracing.FromContext(ctx)

	if k.writer == nil || topic == "" {
		return errPublisherNotConfigured
	}

	start := time.Now()
	err := k.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: message,
			Time:  time.Now(),
		},
	)
	end := time.Since(start)

	if err != nil {
		k.logger.Error("failed to publish message to kafka broker", "error", err)
		return err
	}

	var hostName string

	if len(k.config.Brokers) > 1 {
		hostName = messageMultipleBrokers
	} else {
		hostName = k.config.Brokers[0]
	}

	k.logger.Debug(&pubsub.LogMsg{
		Mode:          "PUB",
		CorrelationID: trackingId,
		MessageValue:  string(message),
		Topic:         topic,
		Host:          hostName,
		Time:          end.Microseconds(),
	})

	return nil
}
