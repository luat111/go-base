// Package redsync provides a distributed lock implementation backed by Redis,
// using the Redlock algorithm via go-redsync/redsync.
//
// Unlike the single-instance lock in pkg/datasource/redis/lock (bsm/redislock),
// redsync supports acquiring locks across multiple independent Redis nodes,
// making it suitable for high-availability deployments.
//
// Basic usage:
//
//	locker := redsync.New(redisClient)
//	mutex, err := locker.Lock(ctx, "my-lock-key", 10*time.Second)
//	if err != nil {
//	    // lock not acquired
//	}
//	defer mutex.Unlock()
package redsync

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"

	redisclient "go-base/pkg/datasource/redis"
)

// DefaultTries is the number of times redsync will attempt to acquire the lock
// before returning ErrFailed.
const DefaultTries = 32

// Locker wraps a redsync.Redsync instance and provides a simplified API for
// obtaining distributed mutexes backed by a single Redis client.
type Locker struct {
	rs *redsync.Redsync
}

// New creates a Locker from an existing *redis.Redis client.
// The underlying redsync pool is built from the client's connection.
func New(client *redisclient.Redis) *Locker {
	pool := goredis.NewPool(client.Client)
	rs := redsync.New(pool)

	return &Locker{rs: rs}
}

// Lock acquires a distributed mutex for the given key with the specified TTL.
// It returns the locked *redsync.Mutex so the caller can extend or release it.
//
// Options can be used to override redsync defaults (e.g. number of retries,
// retry delay). If no options are provided the library defaults apply except
// for the expiry which is always set to ttl.
func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration, opts ...redsync.Option) (*redsync.Mutex, error) {
	defaultOpts := []redsync.Option{
		redsync.WithExpiry(ttl),
		redsync.WithTries(DefaultTries),
	}

	// Caller-supplied opts override defaults.
	allOpts := append(defaultOpts, opts...)

	mutex := l.rs.NewMutex(key, allOpts...)
	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}

	return mutex, nil
}

// TryLock attempts to acquire the distributed mutex exactly once.
// It returns (mutex, nil) on success or (nil, ErrFailed) when the lock is
// already held.
func (l *Locker) TryLock(ctx context.Context, key string, ttl time.Duration) (*redsync.Mutex, error) {
	return l.Lock(ctx, key, ttl, redsync.WithTries(1))
}

// Unlock releases a previously acquired mutex.
// It is safe to call Unlock even if the lock has already expired.
func (l *Locker) Unlock(mutex *redsync.Mutex) (bool, error) {
	return mutex.Unlock()
}

// Extend resets the TTL of an already-held mutex back to the original expiry.
// Use this for long-running tasks that may outlive the initial TTL.
func (l *Locker) Extend(mutex *redsync.Mutex) (bool, error) {
	return mutex.Extend()
}
