package inflight

import (
	"context"
	"go-base/pkg/datasource/redis"
	"time"

	"golang.org/x/sync/singleflight"
)

var group singleflight.Group

func GetOrSet[T any](
	ctx context.Context,
	redisClient *redis.Redis,
	key string,
	getData func() (T, error),
	ttlSeconds int,
) (T, error) {
	var zero T

	// Try Redis cache first
	val, err := redisClient.Get(ctx, key).Result()
	if err == nil && val != "" {
		// You may need to deserialize from string → T here
		// For simplicity, assuming T is string
		return any(val).(T), nil
	}

	// Use singleflight to deduplicate concurrent requests
	result, err, _ := group.Do(key, func() (any, error) {
		data, err := getData()
		if err != nil {
			return zero, err
		}

		// Set to Redis on success
		if ttlSeconds > 0 {
			redisClient.Set(ctx, key, data, time.Duration(ttlSeconds)*time.Second)
		} else {
			redisClient.Set(ctx, key, data, 0)
		}

		return data, nil
	})

	if err != nil {
		return zero, err
	}

	return result.(T), nil
}
