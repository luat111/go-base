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
	rabbitPingTimeout = 5 * time.Second
	defaultRabbitPort = 6379
)

type RabbitConnection struct {
	*amqp091.Connection

	config *RabbitConfig
}

func newRabbitConn(config config.Config, logger logger.ILogger) *RabbitConnection {
	cnf := getConfig(config)
	url := cnf.getConnectionString()

	conn, err := amqp091.Dial(url)

	if err != nil {
		logger.Error("Connect to Rabbit failed", "err", err)
		return nil
	} else {
		logger.Info("Connected to Rabbit")
	}

	return &RabbitConnection{
		Connection: conn,
		config:     cnf,
	}
}

func (c *RabbitConnection) reconnect() error {
	url := c.config.getConnectionString()

	conn, err := amqp091.Dial(url)

	c.Connection = conn

	return err
}

func (c *RabbitConnection) monitorConnection(l logger.ILogger) {
	notifyConnClose := c.Connection.NotifyClose(make(chan *amqp091.Error, 1))

	for {
		select {
		case xErr, ok := <-notifyConnClose:
			if !ok {
				return
			} else {
				l.Error("amqp connection closed", "err", xErr)
				if err := c.reconnect(); err != nil {
					l.Error("amqp connection cannot be reconnected", "err", err)
					return
				}
				notifyConnClose = c.Connection.NotifyClose(make(chan *amqp091.Error, 1))
			}
		}
	}
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
