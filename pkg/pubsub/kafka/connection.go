package kafka

import (
	"context"
	"go-base/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConnection struct {
	conns []*kafka.Conn
}

func setupDialer(conf *Config) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// if conf.SecurityProtocol == protocolSASL || conf.SecurityProtocol == protocolSASLSSL {
	// 	mechanism, err := getSASLMechanism(conf.SASLMechanism, conf.SASLUser, conf.SASLPassword)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	dialer.SASLMechanism = mechanism
	// }

	// if conf.SecurityProtocol == "SSL" || conf.SecurityProtocol == "SASL_SSL" {
	// 	tlsConfig, err := createTLSConfig(&conf.TLS)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	dialer.TLS = tlsConfig
	// }

	return dialer, nil
}

// connectToBrokers connects to Kafka brokers with context support.
func connectToBrokers(ctx context.Context, brokers []string, dialer *kafka.Dialer, logger logger.ILogger) ([]Connection, error) {
	conns := make([]Connection, 0)

	if len(brokers) == 0 {
		return nil, errBrokerNotProvided
	}

	for _, broker := range brokers {
		conn, err := dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			logger.Error("failed to connect to broker %s: %v", broker, err)
			continue
		}

		conns = append(conns, conn)
	}

	if len(conns) == 0 {
		return nil, errFailedToConnectBrokers
	}

	return conns, nil
}

// retryConnect handles the retry mechanism for connecting to the Kafka broker.
func (k *KafkaClient) retryConnect(ctx context.Context) {
	for {
		time.Sleep(defaultRetryTimeout)

		err := k.initialize(ctx)
		if err != nil {
			var brokers any

			if len(k.config.Brokers) > 1 {
				brokers = k.config.Brokers
			} else {
				brokers = k.config.Brokers[0]
			}

			k.logger.Error("could not connect to Kafka at '%v', error: %v", brokers, err)

			continue
		}

		return
	}
}
