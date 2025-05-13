package kafka

import (
	"go-base/pkg/config"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type ReaderConfig struct {
	Brokers        []string
	GroupID        string
	Topic          string
	Client         ClientConfig
	MinBytes       *int
	MaxBytes       int
	CommitInterval *int64
	Key            string
}

func newKafkaReader(c *ReaderConfig, dialer *kafka.Dialer) *kafka.Reader {
	c2 := kafka.ReaderConfig{
		Brokers: c.Brokers,
		GroupID: c.GroupID,
		Topic:   c.Topic,
		Dialer:  dialer,
	}
	if c.CommitInterval != nil {
		c2.CommitInterval = time.Duration(*c.CommitInterval) * time.Nanosecond
	}
	if c.MinBytes != nil && *c.MinBytes >= 0 {
		c2.MinBytes = *c.MinBytes
	}
	if c.MaxBytes > 0 {
		c2.MaxBytes = c.MaxBytes
	}
	return kafka.NewReader(c2)
}

func getReaderConfig(c config.Config) *ReaderConfig {
	var cnf = &ReaderConfig{}

	brokers := strings.Split(c.Get("KAFKA_BROKERS"), ",")
	groupId := c.Get("KAFKA_READER_GROUPID")
	topic := c.Get("KAFKA_READER_TOPIC")
	key := c.Get("KAFKA_READER_KEY")

	cnf.Brokers = brokers
	cnf.GroupID = groupId
	cnf.Topic = topic
	cnf.Key = key

	return cnf
}

func headerToMap(headers []kafka.Header) map[string]string {
	attributes := make(map[string]string, 0)
	for i := range headers {
		attributes[headers[i].Key] = string(headers[i].Value)
	}
	return attributes
}
