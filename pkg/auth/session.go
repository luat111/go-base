package auth

import (
	"context"
	"time"
)

type SessionType string

const (
	SessionSystem            SessionType = "SYSTEM"
	SessionGuestUser         SessionType = "GUEST_USER"
	SessionHybridUser        SessionType = "HYBRID_USER"
	SessionAuthenticatedUser SessionType = "AUTHENTICATED_USER"
)

type SessionStore interface {
	GetByUserAndDevice(ctx context.Context, userID, deviceID string) (*Session, error)
	GetActiveByRefreshToken(ctx context.Context, refreshToken, deviceID string) (*Session, error)
	Upsert(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
}

type KVStore interface {
	Set(ctx context.Context, key, value string, ttlSeconds int) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type Session struct {
	ID           string
	UserID       string
	DeviceID     string
	Secret       string
	RefreshToken string
	ExpireAt     time.Time
	UserAgent    string
	IP           string
}
