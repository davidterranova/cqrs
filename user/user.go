package user

import (
	"github.com/google/uuid"
)

type UserType string

const (
	UserTypeSystem          UserType = "system"
	UserTypeAuthenticated   UserType = "authenticated"
	UserTypeUnauthenticated UserType = "unauthenticated"
)

type UserFactory func() User

type User interface {
	Id() uuid.UUID
	Type() UserType
	String() string
	FromString(string) error
}
