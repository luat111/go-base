package container

import (
	"errors"
	"go-base/pkg/config"
	"go-base/pkg/datasource/postgres"
	"go-base/pkg/datasource/redis"
	"go-base/pkg/kafka"
	"go-base/pkg/logger"
	"go-base/pkg/mq"
	"go-base/pkg/pubsub"
	"reflect"
)

type Container struct {
	appName string

	Logger logger.ILogger

	// metricsManager metrics.Manager
	PubSub pubsub.Client

	Redis *redis.Redis
	DB    *postgres.DB
	MQ    *mq.RabbitClient
	Kafka *kafka.KafkaClient

	// Mongo      Mongo
}

func NewContainer(cnf config.Config) *Container {
	if cnf == nil {
		return &Container{}
	}

	ctn := &Container{
		appName: cnf.Get(config.APP_NAME),
		Logger:  logger.NewLogger(),
	}

	ctn.Create(cnf)

	return ctn
}

func (c *Container) Create(conf config.Config) {
	if c.appName != "" {
		c.appName = conf.GetOrDefault(config.APP_NAME, "go-base")
	}

	c.Logger.Info("Container is being created")

	c.DB = postgres.New(conf, c.Logger)
	c.Redis = redis.New(conf, c.Logger)

	c.initMQ(conf)
	c.initKafka(conf)

	c.PubSub = NewPubsub(conf, c.Logger)
}

func (c *Container) Close() error {
	var err error

	if !isNil(c.DB) {
		err = errors.Join(err, c.DB.Close())
	}

	if !isNil(c.Redis) {
		err = errors.Join(err, c.Redis.Close())
	}

	if !isNil(c.MQ) {
		err = errors.Join(err, c.MQ.Close())
	}

	if !isNil(c.PubSub) {
		err = errors.Join(err, c.PubSub.Close())
	}

	return err
}

func (c *Container) GetAppName() string {
	return c.appName
}

func isNil(i any) bool {
	// Get the value of the interface
	val := reflect.ValueOf(i)

	// If the interface is not assigned or is nil, return true
	return !val.IsValid() || val.IsNil()
}
