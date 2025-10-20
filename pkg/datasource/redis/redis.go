package redis

import (
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
	logger logger.ILogger
	config RedisConfig
}

func New(config config.Config) *Redis {
	logger := logger.NewLogger(common.RedisPrefix)
	client, redisConfig := newRedisClient(config, logger)

	instance := &Redis{
		Client: client,
		logger: logger,
		config: *redisConfig,
	}

	go monitorConnection(instance)

	return instance
}

func (r *Redis) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}

	return nil
}
