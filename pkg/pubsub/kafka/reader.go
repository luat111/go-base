package kafka

import (
	"context"
	"go-base/pkg/pubsub"
	"time"

	"github.com/segmentio/kafka-go"
)

func (k *KafkaClient) createNewReader(topic string) Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     k.config.ConsumerGroupID,
		Brokers:     k.config.Brokers,
		Topic:       topic,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		Dialer:      k.dialer,
		StartOffset: int64(k.config.OffSet),
	})

	return reader
}

func (k *KafkaClient) Subscribe(ctx context.Context, topic string) (*pubsub.Message, error) {
	if !k.isConnected() {
		time.Sleep(defaultRetryTimeout)

		return nil, errClientNotConnected
	}

	if k.config.ConsumerGroupID == "" {
		k.logger.Error("cannot subscribe as consumer_id is not provided in configs")

		return &pubsub.Message{}, ErrConsumerGroupNotProvided
	}

	var reader Reader
	// Lock the reader map to ensure only one subscriber access the reader at a time
	k.mu.Lock()

	if k.reader == nil {
		k.reader = make(map[string]Reader)
	}

	if k.reader[topic] == nil {
		k.reader[topic] = k.createNewReader(topic)
	}

	// Release the lock on the reader map after update
	k.mu.Unlock()

	start := time.Now()

	// Read a single message from the topic
	reader = k.reader[topic]
	msg, err := reader.FetchMessage(ctx)

	if err != nil {
		k.logger.Error("failed to read message from kafka topic", "topic", topic, "error", err)

		return nil, err
	}

	m := pubsub.NewMessage(ctx)
	m.Value = msg.Value
	m.Topic = topic
	m.Committer = newKafkaMessage(&msg, k.reader[topic], k.logger)

	end := time.Since(start)

	var hostName string

	if len(k.config.Brokers) > 1 {
		hostName = "multiple brokers"
	} else {
		hostName = k.config.Brokers[0]
	}

	k.logger.Debug(&pubsub.LogMsg{
		Mode: "SUB",
		// CorrelationID: span.SpanContext().TraceID().String(),
		MessageValue: string(msg.Value),
		Topic:        topic,
		Host:         hostName,
		Time:         end.Microseconds(),
	})

	return m, err
}


