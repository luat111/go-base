package kafka

import (
	"context"
	"errors"
	"go-base/pkg/logger"

	"github.com/segmentio/kafka-go"
)

var (
	errBrokerNotProvided      = errors.New("kafka broker address not provided")
	errFailedToConnectBrokers = errors.New("failed to connect to any kafka brokers")
)

type multiConn struct {
	conns  []*kafka.Conn
	dialer *kafka.Dialer
}

func connectToBrokers(ctx context.Context, brokers []string, dialer *kafka.Dialer, logger logger.ILogger) ([]*kafka.Conn, error) {
	conns := []*kafka.Conn{}

	if len(brokers) == 0 {
		return nil, errBrokerNotProvided
	}

	for _, broker := range brokers {
		conn, err := dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			logger.Error("failed to connect to broker", "broker", broker, "err", err)
			continue
		}

		conns = append(conns, conn)
	}

	if len(conns) == 0 {
		return nil, errFailedToConnectBrokers
	}

	return conns, nil
}
