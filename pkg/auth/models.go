package auth

import "time"


type RequestContext struct {
	DeviceID  string
	UserAgent string
	IP        string
}

type LoginPayload[T any] struct {
	Credential T
}


type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpireAt     time.Time
}
