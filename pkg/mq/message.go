package mq

type Message struct {
	ExchangeName string
	Route        string
	Body         any
	Headers      map[string]string
}
