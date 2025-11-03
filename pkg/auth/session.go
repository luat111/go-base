package auth

type IBaseSessionEntity interface {
	Id() string
	UserId() string
}