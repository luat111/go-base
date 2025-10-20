package kafka

import (
	"context"
	"errors"
	"go-base/pkg/common"
	"go-base/pkg/logger"
	"sync"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	dialer *kafka.Dialer
	conn   *multiConn

	writer Writer
	reader map[string]Reader

	mu *sync.RWMutex

	logger logger.ILogger
	config Config
}

func New(conf *Config) *KafkaClient {
	logger := logger.NewLogger(common.PubsubPrefix)
	// returning unexported types as intended.
	err := validateConfigs(conf)
	if err != nil {
		logger.Error("could not initialize kafka", "error", err)

		return nil
	}

	if len(conf.Brokers) == 1 {
		logger.Log("connecting to Kafka broker:", "broker", conf.Brokers[0])
	} else {
		logger.Log("connecting to Kafka brokers", "brokers", conf.Brokers)
	}

	client := &KafkaClient{
		logger: logger,
		config: *conf,
		mu:     &sync.RWMutex{},
	}
	ctx := context.Background()

	err = client.initialize(ctx)

	if err != nil {
		logger.Error("failed to connect to kafka at %v, error: %v", conf.Brokers, err)

		go client.retryConnect(ctx)

		return client
	}

	return client
}

func (k *KafkaClient) initialize(ctx context.Context) error {
	dialer, err := setupDialer(&k.config)
	if err != nil {
		return err
	}

	conns, err := connectToBrokers(ctx, k.config.Brokers, dialer, k.logger)
	if err != nil {
		return err
	}

	multi := &multiConn{
		conns:  conns,
		dialer: dialer,
	}

	writer := createKafkaWriter(&k.config, dialer, k.logger)
	reader := make(map[string]Reader)

	k.logger.Log("connected to %d Kafka brokers", len(conns))

	k.dialer = dialer
	k.conn = multi
	k.writer = writer
	k.reader = reader

	return nil
}

func (k *KafkaClient) isConnected() bool {
	if k.conn == nil {
		return false
	}

	_, err := k.conn.Controller()

	return err == nil
}

func (k *KafkaClient) Close() (err error) {
	for _, r := range k.reader {
		err = errors.Join(err, r.Close())
	}

	if k.writer != nil {
		err = errors.Join(err, k.writer.Close())
	}

	if k.conn != nil {
		err = errors.Join(k.conn.Close())
	}

	return err
}

func (k *KafkaClient) DeleteTopic(_ context.Context, name string) error {
	return k.conn.DeleteTopics(name)
}

func (k *KafkaClient) Controller() (broker kafka.Broker, err error) {
	return k.conn.Controller()
}

func (k *KafkaClient) CreateTopic(_ context.Context, name string) error {
	topics := kafka.TopicConfig{Topic: name, NumPartitions: 1, ReplicationFactor: 1}

	return k.conn.CreateTopics(topics)
}
