package mq

import (
	"go.opentelemetry.io/otel/propagation"
)

// amqpHeaderCarrier adapts an amqp091.Table so it satisfies
// propagation.TextMapCarrier. This lets the global TextMapPropagator
// inject/extract W3C TraceContext (and Baggage) headers directly
// into/from RabbitMQ message headers.
type amqpHeaderCarrier map[string]string

var _ propagation.TextMapCarrier = amqpHeaderCarrier{}

func (c amqpHeaderCarrier) Get(key string) string { return c[key] }
func (c amqpHeaderCarrier) Set(key, val string)   { c[key] = val }
func (c amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
