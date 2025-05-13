package kafka

import (
	"context"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"strings"

	"github.com/segmentio/kafka-go/sasl/scram"
)

type HandlerFunc func(context.Context, []byte, map[string]string)

type KafkaClient struct {
	*Consumer
	*Producer

	conn   *multiConn
	Logger logger.ILogger
}

func New(conf config.Config, log logger.ILogger) *KafkaClient {
	client := &KafkaClient{}

	err := client.initialize(conf, log)

	if err != nil {
		return nil
	}

	return client
}

func (k *KafkaClient) initialize(conf config.Config, log logger.ILogger) error {
	clientConf := getConfig(conf)
	dialer := getDialer(clientConf, scram.SHA512, nil)

	brokers := strings.Split(conf.Get("KAFKA_BROKERS"), ",")
	conns, err := connectToBrokers(context.Background(), brokers, dialer, log)
	if err != nil {
		return err
	}

	multi := &multiConn{
		conns:  conns,
		dialer: dialer,
	}

	k.conn = multi
	k.Logger = log

	k.Logger.Log("connected to Kafka brokers")

	return nil
}

func (k *KafkaClient) InitProducer(conf config.Config) error {
	producer, err := NewProducer(k.conn, conf, k.Logger)

	if err != nil {
		return err
	}

	k.Producer = producer

	k.Logger.Log("Kafka Producer initialized")

	return nil
}

func (k *KafkaClient) InitConsumer(conf config.Config, autoAck bool) error {
	consumer, err := NewConsumer(k.conn, conf, k.Logger, autoAck)

	if err != nil {
		return err
	}

	k.Consumer = consumer

	k.Logger.Log("Kafka Consumer initialized")

	return nil
}
