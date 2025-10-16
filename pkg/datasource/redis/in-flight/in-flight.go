package inflight

import (
	"context"
	"go-base/pkg/datasource/redis"
	"sync"
	"time"
)

var callingMap sync.Map

type callEntry struct {
	wg  sync.WaitGroup
	val any
	err error
	mu  sync.Mutex
}

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

	// Check if the request is already in-flight
	entryIface, loaded := callingMap.LoadOrStore(key, &callEntry{})
	entry := entryIface.(*callEntry)

	if !loaded {
		// This goroutine is responsible for fetching
		entry.wg.Add(1)
		defer func() {
			callingMap.Delete(key)
			entry.wg.Done()
		}()

		var result T

		entry.mu.Lock()

		result, entry.err = getData()
		entry.val = result

		entry.mu.Unlock()

		if entry.err == nil {
			// Set to Redis
			if ttlSeconds > 0 {
				redisClient.Set(ctx, key, result, time.Duration(ttlSeconds)*time.Second)
			} else {
				redisClient.Set(ctx, key, result, 0)
			}
		}
	} else {
		// Wait for the in-flight call to finish
		entry.wg.Wait()
	}

	if entry.err != nil {
		return zero, entry.err
	}

	return entry.val.(T), nil
}
