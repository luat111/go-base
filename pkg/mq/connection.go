package mq

import (
	"fmt"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	defaultRabbitPort               = 5672
	timeOutRetry      time.Duration = 5 * time.Second
	timeOutDuration   time.Duration = 5 * time.Second
)

type RabbitConnection struct {
	*amqp091.Connection

	config *RabbitConfig
}

func newRabbitConn(config config.Config, logger logger.ILogger) (*RabbitConnection, error) {
	cnf := getConfig(config)
	url := cnf.getConnectionString()

	conn, err := amqp091.Dial(url)

	if err != nil {
		logger.Error("Connect to Rabbit failed", "err", err)
		panic(err)
	} else {
		logger.Info("Connected to Rabbit")
	}

	return &RabbitConnection{
		Connection: conn,
		config:     cnf,
	}, nil
}

func (c *RabbitConnection) reconnect() error {
	url := c.config.getConnectionString()

	conn, err := amqp091.Dial(url)

	c.Connection = conn

	return err
}

type RabbitConfig struct {
	HostName string
	Username string
	Password string
	Port     int
}

func getConfig(c config.Config) *RabbitConfig {
	var cnf = &RabbitConfig{}

	cnf.HostName = c.Get("RMQ_HOST")
	cnf.Username = c.Get("RMQ_USER")

	cnf.Password = c.Get("RMQ_PWD")

	port, err := strconv.Atoi(c.Get("RMQ_PORT"))
	if err != nil {
		port = defaultRabbitPort
	}

	cnf.Port = port

	return cnf
}

func (c *RabbitConfig) getConnectionString() string {
	connectString := fmt.Sprintf(
		"amqp://%s:%s@%s:%d",
		c.Username,
		c.Password,
		c.HostName,
		c.Port,
	)

	return connectString
}
