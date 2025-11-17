package auth

import "context"


type User interface {
	GetID() string
}

type UserRepository[U User] interface {
	FindOne(ctx context.Context, cond any) (U, error)
}
