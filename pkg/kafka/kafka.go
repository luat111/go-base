package kafka

import (
	"context"
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"strings"

	"github.com/segmentio/kafka-go/sasl/scram"
)

type HandlerFunc func(context.Context, []byte, map[string]string)

type KafkaClient struct {
	consumers map[string]*Consumer
	*Producer

	conn   *multiConn
	Logger logger.ILogger
}

func New(conf config.Config) *KafkaClient {
	log := logger.NewLogger(common.KafkaPrefix)
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

func (k *KafkaClient) InitConsumer(conf config.Config, topic string, handler HandlerFunc, autoAck bool) error {
	consumer, err := NewConsumer(k.conn, conf, k.Logger, topic, handler, autoAck)

	if err != nil {
		return err
	}

	k.consumers[topic] = consumer

	k.Logger.Log("Kafka Consumer initialized")

	return nil
}

func (k *KafkaClient) StartConsume() {
	for _, consumer := range k.consumers {
		consumer.Consume()
	}

	k.Logger.Log("Consuming kafka message")
}
