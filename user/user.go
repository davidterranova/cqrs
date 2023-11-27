package user

import (
	"github.com/google/uuid"
)

type UserFactory func() User

type User interface {
	Id() uuid.UUID
	String() string
	FromString(string) error
}
