package auth

import "time"

type RequestContext struct {
	UserID      string
	DeviceID    string
	UserAgent   string
	IP          string
	SessionType SessionType
	Allowed     []string
}

type LoginPayload[T any] struct {
	Credential T
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpireAt     time.Time
}
