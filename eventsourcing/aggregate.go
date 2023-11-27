package eventsourcing

import (
	"time"

	"github.com/google/uuid"
)

type AggregateType string

type Aggregate interface {
	AggregateId() uuid.UUID
	AggregateType() AggregateType
	AggregateVersion() int

	IncrementVersion()
}

type AggregateBase[T Aggregate] struct {
	aggregateId      uuid.UUID
	aggregateVersion int
	events           []Event[T]

	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewAggregateBase[T Aggregate](aggregateId uuid.UUID, version int) *AggregateBase[T] {
	now := time.Now().UTC()
	return &AggregateBase[T]{
		aggregateId:      aggregateId,
		aggregateVersion: version,
		createdAt:        now,
		updatedAt:        now,
		events:           make([]Event[T], 0),
	}
}

func (a AggregateBase[T]) AggregateId() uuid.UUID {
	return a.aggregateId
}

func (a *AggregateBase[T]) Init(e Event[T]) {
	a.aggregateId = e.AggregateId()
	a.createdAt = e.IssuedAt()
	a.Process(e)
}

func (a *AggregateBase[T]) Delete(e Event[T]) {
	now := e.IssuedAt()
	a.deletedAt = &now
	a.Process(e)
}

func (a *AggregateBase[T]) Process(e Event[T]) {
	a.aggregateVersion = e.AggregateVersion()
	a.events = append(a.events, e)
	a.updatedAt = e.IssuedAt()
}

func (a *AggregateBase[T]) IncrementVersion() {
	a.aggregateVersion++
}

func (a AggregateBase[T]) AggregateVersion() int {
	return a.aggregateVersion
}

func (a AggregateBase[T]) Events() []Event[T] {
	return a.events
}

func (a AggregateBase[T]) CreatedAt() time.Time {
	return a.createdAt
}

func (a AggregateBase[T]) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a AggregateBase[T]) DeletedAt() *time.Time {
	return a.deletedAt
}
