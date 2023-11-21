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

type User interface {
	Id() uuid.UUID
	Type() UserType
	IsAuthenticatedOrSystem() bool
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	String() string
	FromString(string) error
}
