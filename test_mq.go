package main

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func TestMQ(body []byte, metadata map[string]string, msg amqp091.Delivery) {
	fmt.Println(body, metadata)
}
