package mq

import (
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/logger"

	"golang.org/x/sync/errgroup"
)

type RabbitClient struct {
	*Consumer
	*Producer

	conn   *RabbitConnection
	Logger logger.ILogger
}

func New(config config.Config) *RabbitClient {
	logger := logger.NewLogger(common.MQPrefix)
	conn := newRabbitConn(config, logger)

	go conn.monitorConnection(logger)

	client := &RabbitClient{
		conn:   conn,
		Logger: logger,
	}

	return client
}

func (r *RabbitClient) Init(appName string, autoAck bool) {
	exchangeName := generateExchangeName(appName)
	queueName := generateQueueName(appName)

	consumer := newConsumer(r, exchangeName, queueName, autoAck)
	producer := newProducer(r)

	r.Consumer, r.Producer = consumer, producer
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
