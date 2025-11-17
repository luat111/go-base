package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTPayload interface {
	GetID() string
	GetDeviceID() string
	GetType() SessionType
	GetAllowed() []string

	// Optional helpers similar to BaseJWTPayload
	GetIssuedAt() time.Time
	GetExpiresAt() time.Time
}


type JWTSigner interface {
	Sign(payload JWTPayload, secret string, ttl time.Duration) (string, error)
	Verify(tokenStr, secret string) (JWTPayload, error)
	Decode(tokenStr string) (JWTPayload, error)
}

type defaultJWTSigner struct {
	keyFunc func(claims jwt.MapClaims) JWTPayload
}

func (s *defaultJWTSigner) Sign(payload JWTPayload, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"id":       payload.GetID(),
		"deviceId": payload.GetDeviceID(),
		"type":     string(payload.GetType()),
		"allowed":  payload.GetAllowed(),
		"iat":      now.Unix(),
		"exp":      now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *defaultJWTSigner) Verify(tokenStr, secret string) (JWTPayload, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return s.keyFunc(mc), nil
}

func (s *defaultJWTSigner) Decode(tokenStr string) (JWTPayload, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return s.keyFunc(mc), nil
}
