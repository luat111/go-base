package kafka

import (
	"go-base/pkg/config"
	"strings"

	"github.com/segmentio/kafka-go"
)

type WriterConfig struct {
	Brokers []string
	Topic   string
	Client  ClientConfig
}

func newKafkaWriter(conf *WriterConfig, dialer *kafka.Dialer) *kafka.Writer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  conf.Brokers,
		Topic:    conf.Topic,
		Dialer:   dialer,
		Balancer: &kafka.LeastBytes{},
	})
	return writer
}

func getWriterConfig(c config.Config) *WriterConfig {
	var cnf = &WriterConfig{}

	brokers := strings.Split(c.Get("KAFKA_BROKERS"), ",")
	topic := c.Get("KAFKA_WRITER_TOPIC")

	cnf.Brokers = brokers
	cnf.Topic = topic

	return cnf
}

func mapToHeader(attributes map[string]string) []kafka.Header {
	headers := make([]kafka.Header, 0)
	for k, v := range attributes {
		h := kafka.Header{Key: k, Value: []byte(v)}
		headers = append(headers, h)
	}
	return headers
}
