package mq

import (
	"cmp"
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

type RabbitClient struct {
	*Consumer
	*Producer

	conn   *RabbitConnection
	Logger logger.ILogger

	appName string
	autoAck bool

	handlers map[string]HandlerFunc
}

func New(config config.Config) (*RabbitClient, error) {
	logger := logger.NewLogger(common.MQPrefix)
	conn, err := newRabbitConn(config, logger)

	client := &RabbitClient{
		conn:   conn,
		Logger: logger,
	}

	go client.monitorConnection(logger)

	return client, err
}

func (r *RabbitClient) Init(appName string, autoAck bool) error {
	exchangeName := generateExchangeName(appName)
	queueName := generateQueueName(appName)

	if r.conn == nil {
		r.Logger.Error("Rabbit connection is nil")
		return errClientConnIsNil
	}

	consumer := newConsumer(r, exchangeName, queueName, autoAck)
	producer := newProducer(r)

	r.Consumer, r.Producer = consumer, producer
	r.appName, r.autoAck = appName, autoAck

	return nil
}

func (r *RabbitClient) Close() error {
	g := new(errgroup.Group)

	g.Go(func() error {
		if r.Consumer == nil {
			return nil
		}

		return r.Consumer.Channel.Close()
	})

	g.Go(func() error {
		if r.Producer == nil || r.Producer.Channel == nil {
			return nil
		}

		return r.Producer.Channel.Close()
	})

	if err := g.Wait(); err != nil {
		r.Logger.Error("Close rabbit channel failed", "err", err)
		return err
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.Logger.Error("Close rabbit connection failed", "err", err)
			return err
		}
	}

	return nil
}

func (r *RabbitClient) BindQueue(route string, handler HandlerFunc) error {
	if r.Consumer == nil {
		return errConsumerIsNil
	}
	return r.Consumer.BindQueue(route, handler)
}

func (r *RabbitClient) Listen(handlers map[string]HandlerFunc) error {
	if r == nil {
		return errClientIsNil
	}

	g := new(errgroup.Group)

	for route, handler := range handlers {
		g.Go(func() error {
			return r.BindQueue(route, handler)
		})
	}

	if err := g.Wait(); err != nil {
		r.Logger.Error("Listen RMQ failed", "err", err)
		return err
	}

	r.handlers = handlers
	r.Consume()

	return nil
}

func (r *RabbitClient) reInit() error {
	errInit := r.Init(r.appName, r.autoAck)
	errConsume := r.Listen(r.handlers)

	return cmp.Or(errInit, errConsume)
}

func (c *RabbitClient) monitorConnection(l logger.ILogger) {
	if c == nil || c.conn == nil {
		l.Error("RabbitConnection is nil, cannot monitor")
		return
	}

	for {

		reason, ok := <-c.conn.NotifyClose(make(chan *amqp091.Error, 1))
		if !ok {
			l.Info("Rabbitmq connection closed")
			break
		}

		l.Warn("Rabbitmq connection closed unexpectedly", "reason", reason)

		for {

			if err := c.conn.reconnect(); err != nil {
				l.Error("Rabbitmq reconnect failed", "err", err)
			} else {
				if errInit := c.reInit(); errInit == nil {
					l.Info("Rabbitmq reconnect success")
					break
				}
			}

			time.Sleep(timeOutRetry)
		}

		time.Sleep(timeOutDuration)
	}
}
