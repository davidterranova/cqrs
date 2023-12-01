package eventsourcing

import (
	"time"

	"github.com/google/uuid"
)

type Command[T Aggregate] interface {
	// AggregateId returns the id of the aggregate on which the command should be applied
	AggregateId() uuid.UUID
	// AggregateType returns the type of the aggregate on which the command should be applied
	AggregateType() AggregateType
	// CreatedAt returns the time at which the command was created
	CreatedAt() time.Time
	// IssuedBy returns the user who issued the command
	IssuedBy() User

	// Check for validity of command on aggregate, mutate the aggregate and return newly emitted events
	Apply(*T) ([]Event[T], error)
}

type BaseCommand[T Aggregate] struct {
	BCAggregateId   uuid.UUID     `validate:"required"`
	BCAggregateType AggregateType `validate:"required"`
	BCCreatedAt     time.Time     `validate:"required"`
	BCIssuedBy      User          `validate:"required"`
}

func NewBaseCommand[T Aggregate](aggregateId uuid.UUID, aggregateType AggregateType, issuedBy User) BaseCommand[T] {
	return BaseCommand[T]{
		BCAggregateId:   aggregateId,
		BCAggregateType: aggregateType,
		BCIssuedBy:      issuedBy,
		BCCreatedAt:     time.Now().UTC(),
	}
}

func (c BaseCommand[T]) AggregateId() uuid.UUID {
	return c.BCAggregateId
}

func (c BaseCommand[T]) AggregateType() AggregateType {
	return c.BCAggregateType
}

func (c BaseCommand[T]) CreatedAt() time.Time {
	return c.BCCreatedAt
}

func (c BaseCommand[T]) IssuedBy() User {
	return c.BCIssuedBy
}
