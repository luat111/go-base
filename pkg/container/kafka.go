package container

import (
	"go-base/pkg/config"
	"go-base/pkg/kafka"
	"strconv"
)

func (c *Container) initKafka(conf config.Config) {
	kkUser := conf.Get(config.KAFKA_USER)
	kkPwd := conf.Get(config.KAFKA_PWD)
	if kkUser != "" && kkPwd != "" {
		kkClient := kafka.New(conf, c.Logger)
		autoAck, errAck := strconv.ParseBool(conf.Get(config.KAFKA_ACK))

		if errAck != nil {
			autoAck = true
		}

		c.Kafka = kkClient

		reader := conf.Get(config.KAFKA_READER_TOPIC)
		if reader != "" {
			c.Kafka.InitConsumer(conf, autoAck)
		}

		writer := conf.Get(config.KAFKA_WRITER_TOPIC)
		if writer != "" {
			c.Kafka.InitProducer(conf)
		}
	}
}
