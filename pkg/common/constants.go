package common

import "time"

const (
	CheckPortTimeout time.Duration = 2 * time.Second
	DefaultTimeOut   time.Duration = 10 * time.Second
	DefaultHTTPPort  int           = 3000
	ReqParams        string        = "req_params"
)

const (
	AppPrefix = "APP"
	ContainerPrefix = "CONTAINER"
	PGPrefix = "POSTGRES"
	RedisPrefix = "REDIS"
	PubsubPrefix = "PUBSUB"
	MQPrefix = "MQ"
	KafkaPrefix = "KAFKA"
	CronPrefix = "CRON"
	RPCPrefix = "GRPC"
	HTTPPrefix = "HTTP"
)