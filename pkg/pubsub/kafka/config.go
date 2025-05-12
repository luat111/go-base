package kafka

import "time"

type Config struct {
	Brokers          []string
	Partition        int
	ConsumerGroupID  string
	OffSet           int
	BatchSize        int
	BatchBytes       int
	BatchTimeout     int
	RetryTimeout     time.Duration
	SASLMechanism    string
	SASLUser         string
	SASLPassword     string
	SecurityProtocol string
	// TLS              TLSConfig
}