package eventsourcing

import (
	"errors"

	"github.com/google/uuid"
)

type UserFactory func() User

type User interface {
	Id() uuid.UUID
	String() string
	FromString(string) error
}

var (
	ErrInvalidUser = errors.New("invalid user")
	SystemUser     = &systemUser{}

	systemUserId = uuid.MustParse("99999999-9999-9999-9999-999999999999")
)

type systemUser struct{}

func (u systemUser) Id() uuid.UUID  { return systemUserId }
func (u systemUser) String() string { return "system" }
func (u systemUser) FromString(s string) error {
	if s != "system" {
		return ErrInvalidUser
	}

	return nil
}
