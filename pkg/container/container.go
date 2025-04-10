package container

import (
	"errors"
	"go-base/pkg/config"
	"go-base/pkg/datasource/postgres"
	"go-base/pkg/datasource/redis"
	"go-base/pkg/logger"
	"go-base/pkg/mq"
	"reflect"
	"strconv"
)

type Container struct {
	appName string

	Logger logger.ILogger

	// metricsManager metrics.Manager
	// PubSub         pubsub.Client

	Redis *redis.Redis
	DB    *postgres.DB
	MQ    *mq.RabbitClient

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

	mqHost := conf.Get(config.RMQ_HOST)
	if mqHost != "" {
		c.MQ = mq.New(conf, c.Logger)
		autoAck, err := strconv.ParseBool(conf.Get(config.RMQ_ACK))

		if err != nil {
			autoAck = true
		}

		c.MQ.Init(c.appName, autoAck)
	}
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
