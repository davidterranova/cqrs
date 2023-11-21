package eventsourcing

import (
	"fmt"
	"time"

	"github.com/davidterranova/cqrs/user"

	"github.com/google/uuid"
)

type Event[T Aggregate] interface {
	Id() uuid.UUID
	AggregateId() uuid.UUID
	AggregateType() AggregateType
	EventType() string
	IssuedAt() time.Time
	IssuedBy() user.User
	Apply(*T) error

	// SetBase(EventBase[T]) is used internally by eventsourcing package
	SetBase(EventBase[T])
	AggregateVersion() int
}

type EventBase[T Aggregate] struct {
	eventId          uuid.UUID
	eventIssuesAt    time.Time
	eventIssuedBy    user.User
	eventType        string
	aggregateType    AggregateType
	aggregateId      uuid.UUID
	aggregateVersion int
}

func NewEventBase[T Aggregate](aggregateType AggregateType, aggregateVersion int, eventType string, aggregateId uuid.UUID, issuedBy user.User) *EventBase[T] {
	return &EventBase[T]{
		eventId:          uuid.New(),
		eventIssuedBy:    issuedBy,
		eventIssuesAt:    time.Now().UTC(),
		eventType:        eventType,
		aggregateType:    aggregateType,
		aggregateId:      aggregateId,
		aggregateVersion: aggregateVersion,
	}
}

func NewEventBaseFromRepository[T Aggregate](eventId uuid.UUID, eventType string, aggregateType AggregateType, aggregateId uuid.UUID, aggregateVersion int, issuedBy user.User) *EventBase[T] {
	return &EventBase[T]{
		eventId:          eventId,
		eventIssuedBy:    issuedBy,
		eventIssuesAt:    time.Now().UTC(),
		eventType:        eventType,
		aggregateType:    aggregateType,
		aggregateId:      aggregateId,
		aggregateVersion: aggregateVersion,
	}
}

func (e EventBase[T]) Id() uuid.UUID {
	return e.eventId
}

func (e EventBase[T]) AggregateId() uuid.UUID {
	return e.aggregateId
}

func (e EventBase[T]) IssuedAt() time.Time {
	return e.eventIssuesAt
}

func (e EventBase[T]) AggregateType() AggregateType {
	return e.aggregateType
}

func (e EventBase[T]) IssuedBy() user.User {
	return e.eventIssuedBy
}

func (e EventBase[T]) EventType() string {
	return e.eventType
}

func (e EventBase[T]) AggregateVersion() int {
	return e.aggregateVersion
}

func (e *EventBase[T]) SetBase(base EventBase[T]) {
	e.eventId = base.eventId
	e.eventIssuesAt = base.eventIssuesAt
	e.eventIssuedBy = base.eventIssuedBy
	e.eventType = base.eventType
	e.aggregateType = base.aggregateType
	e.aggregateId = base.aggregateId
	e.aggregateVersion = base.aggregateVersion
}

func (e EventBase[T]) String() string {
	return fmt.Sprintf("#%s by:%s at:%s %s.%s on:%s", e.eventId, e.eventIssuedBy, e.eventIssuesAt, e.aggregateType, e.eventType, e.aggregateId)
}
