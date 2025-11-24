package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Authenticator interface {
	Login(ctx context.Context, payload LoginPayload[any], rc RequestContext) (TokenPair, error)
	Logout(ctx context.Context, userID string, rc RequestContext) error
	Refresh(ctx context.Context, refreshToken string, rc RequestContext) (TokenPair, error)
}

type CredentialsValidator[U User, C any] interface {
	Validate(ctx context.Context, user U, cred C) error
}

type BaseJWTAuth[U User, C any] struct {
	store         SessionStore
	kv            KVStore
	users         UserRepository[U]
	validator     CredentialsValidator[U, C]
	signer        JWTSigner
	defaultExpire time.Duration
}

func NewBaseJWTAuth[U User, C any](
	store SessionStore,
	kv KVStore,
	users UserRepository[U],
	validator CredentialsValidator[U, C],
	signer JWTSigner,
	defaultExpire time.Duration,
) *BaseJWTAuth[U, C] {
	if signer == nil {
		signer = &defaultJWTSigner{ /* inject claim -> payload func */ }
	}
	if defaultExpire == 0 {
		defaultExpire = time.Minute * 15
	}
	return &BaseJWTAuth[U, C]{store, kv, users, validator, signer, defaultExpire}
}

func (a *BaseJWTAuth[U, C]) genSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (a *BaseJWTAuth[U, C]) Login(ctx context.Context, payload LoginPayload[any], rc RequestContext) (TokenPair, error) {
	cond := payload.Credential // up to your repo
	u, err := a.users.FindOne(ctx, cond)
	if err != nil {
		return TokenPair{}, err
	}
	if err := a.validator.Validate(ctx, u, payload.Credential.(C)); err != nil {
		return TokenPair{}, err
	}
	return a.handleSession(ctx, u, rc)
}

func (a *BaseJWTAuth[U, C]) handleSession(ctx context.Context, u U, rc RequestContext) (TokenPair, error) {
	secret, err := a.genSecret()
	if err != nil {
		return TokenPair{}, err
	}

	// build your own JWTPayload implementation here
	payload := newJwtPayload(rc.UserID, rc.DeviceID, rc.SessionType, rc.Allowed) /* your concrete JWTPayload for the user */

	access, err := a.signer.Sign(payload, secret, a.defaultExpire)
	if err != nil {
		return TokenPair{}, err
	}

	refreshSecret, _ := a.genSecret()
	refresh, err := a.signer.Sign(payload, refreshSecret, a.defaultExpire*4)
	if err != nil {
		return TokenPair{}, err
	}

	exp := time.Now().Add(a.defaultExpire)
	session := &Session{
		UserID:       u.GetID(),
		DeviceID:     rc.DeviceID,
		Secret:       secret,
		RefreshToken: refresh,
		ExpireAt:     exp,
		UserAgent:    rc.UserAgent,
		IP:           rc.IP,
	}
	if err := a.store.Upsert(ctx, session); err != nil {
		return TokenPair{}, err
	}

	// also store secret in KV if you need a hash-based key like ParserHelper.genHashUserAgent
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpireAt:     exp,
	}, nil
}
