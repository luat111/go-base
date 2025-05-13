package container

import (
	"go-base/pkg/config"
	"go-base/pkg/kafka"
)

func (c *Container) initKafka(conf config.Config) {
	kkUser := conf.Get(config.KAFKA_USER)
	kkPwd := conf.Get(config.KAFKA_PWD)
	if kkUser != "" && kkPwd != "" {
		kkClient := kafka.New(conf, c.Logger)
		c.Kafka = kkClient
	}
}
