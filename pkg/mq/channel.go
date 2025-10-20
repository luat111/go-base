package mq

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	timeOutRetry    time.Duration = 5 * time.Second
	timeOutDuration time.Duration = 30 * time.Second
)

type Channel struct {
	*amqp091.Channel

	client *RabbitClient
}

func newChannel(client *RabbitClient) (*Channel, error) {
	if client == nil || client.conn == nil {
		return nil, errClientConnIsNil
	}

	channel, err := client.conn.Channel()

	if err != nil {
		client.Logger.Error("Create channel failed", "err", err)
		return nil, err
	}

	ch := &Channel{Channel: channel, client: client}

	go ch.monitorChannel()

	return ch, nil
}

func (c *Channel) monitorChannel() *Channel {
	notifyChannelClose := c.NotifyClose(make(chan *amqp091.Error, 1))
	for {
		xErr, ok := <-notifyChannelClose
		if !ok {
			return nil
		} else {
			log.Printf("AMQP channel connection lost: %v. Attempting to reconnect...", xErr)
			for retries := 1; retries <= 10; retries++ {
				log.Printf("Reconnection attempt #%d...", retries)

				if newChannel, err := newChannel(c.client); err == nil {
					log.Println("Reconnected successfully.")
					return newChannel
				}

				time.Sleep(timeOutRetry)
			}
		}

		time.Sleep(timeOutDuration)
	}
}
