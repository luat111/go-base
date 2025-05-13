package lock

import (
	"context"
	"go-base/pkg/datasource/redis"
	"time"

	"github.com/bsm/redislock"
)

type Locker struct {
	*redislock.Client
	redisClient *redis.Redis
}

func New(client *redis.Redis) *Locker {
	locker := redislock.New(client)

	return &Locker{
		Client:      locker,
		redisClient: client,
	}
}

func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration) (*redislock.Lock, error) {
	return l.Client.Obtain(ctx, key, ttl, nil)
}
