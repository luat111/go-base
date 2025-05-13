package kafka

import (
	"go-base/pkg/config"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type ClientConfig struct {
	Username string
	Password string
	Timeout  int64
}

func getConfig(c config.Config) *ClientConfig {
	var cnf = &ClientConfig{}

	cnf.Username = c.Get("KAFKA_USER")
	cnf.Password = c.Get("KAFKA_PWD")

	timeout, err := strconv.ParseInt(c.Get("KAFKA_TIMEOUT"), 10, 64)
	if err != nil {
		timeout = 5000
	}

	cnf.Timeout = timeout

	return cnf
}

func getDialer(cnf *ClientConfig, algorithm scram.Algorithm, dialer *kafka.Dialer) *kafka.Dialer {
	if dialer == nil {
		dialer = &kafka.Dialer{
			Timeout:   1 * time.Minute,
			DualStack: true,
		}
	}
	if cnf.Username != "" && cnf.Password != "" {
		mechanism, err := scram.Mechanism(algorithm, cnf.Username, cnf.Password)
		if err != nil {
			panic(err)
		}
		dialer.SASLMechanism = mechanism
	}
	return dialer
}
