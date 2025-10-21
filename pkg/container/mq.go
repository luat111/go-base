package container

import (
	"go-base/pkg/config"
	"go-base/pkg/mq"
	"strconv"
)

func (c *Container) initMQ(conf config.Config) {
	mqHost := conf.Get(config.RMQ_HOST)
	if mqHost != "" {
		if mq, err := mq.New(conf); err == nil {
			c.MQ = mq
		}

		autoAck, err := strconv.ParseBool(conf.Get(config.RMQ_ACK))
		if err != nil {
			autoAck = true
		}

		c.MQ.Init(c.appName, autoAck)
	}
}
