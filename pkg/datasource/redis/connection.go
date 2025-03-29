package redis

import (
	"context"
	"fmt"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisPingTimeout = 5 * time.Second
	defaultRedisPort = 6379
)

type RedisConfig struct {
	HostName string
	Username string
	Password string
	Port     int
	DB       int
	Options  *redis.Options
}

func newRedisClient(config config.Config, logger logger.ILogger) (*redis.Client, *RedisConfig) {
	redisConfig := getRedisConfig(config)

	rc := redis.NewClient(redisConfig.Options)
	// Add logger and metric for redis
	// rc.AddHook(&redisHook{config: redisConfig, logger: logger, metrics: metrics})

	if err := checkConnection(rc); err != nil {
		logger.Error("Failed to connect to the Redis!", "err", err.Error())
	} else {
		logger.Info("Connected to Redis")
	}

	return rc, redisConfig
}

func getRedisConfig(c config.Config) *RedisConfig {
	var redisConfig = &RedisConfig{}

	redisConfig.HostName = c.Get("CACHE_HOST")

	redisConfig.Password = c.Get("CACHE_PWD")

	port, err := strconv.Atoi(c.Get("CACHE_PORT"))
	if err != nil {
		port = defaultRedisPort
	}

	redisConfig.Port = port

	db, err := strconv.Atoi(c.Get("CACHE_DB"))
	if err != nil {
		db = 0 // default to DB 0 if not specified
	}

	redisConfig.DB = db

	options := new(redis.Options)

	if options.Addr == "" {
		options.Addr = fmt.Sprintf("%s:%d", redisConfig.HostName, redisConfig.Port)
	}

	if options.Password == "" {
		options.Password = redisConfig.Password
	}

	options.DB = redisConfig.DB

	redisConfig.Options = options

	return redisConfig
}

func monitorConnection(r *Redis) {
	const connRetryFrequencyInSeconds = 10
	for {
		err := checkConnection(r.Client)
		if err != nil {
			r.logger.Warn("Connection to redis lost")

			for {
				r.logger.Warn("Retrying connect to redis")

				r.Client = redis.NewClient(r.config.Options)

				if err := checkConnection(r.Client); err == nil {
					r.logger.Info("Reconnected to redis")
					break
				}

				time.Sleep(connRetryFrequencyInSeconds * time.Second)
			}
		}

		time.Sleep(connRetryFrequencyInSeconds * time.Second)
	}
}

func checkConnection(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	}

	return nil
}
