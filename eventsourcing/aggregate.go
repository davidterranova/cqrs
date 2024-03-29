package eventsourcing

import (
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type AggregateType string

type Aggregate interface {
	AggregateId() uuid.UUID
	AggregateType() AggregateType
	AggregateVersion() int
}

type AggregateBase[T Aggregate] struct {
	aggregateId      uuid.UUID
	aggregateVersion int
	events           []Event[T]

	issuedBy  User
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

func NewFullAggregateBase[T Aggregate](aggregateId uuid.UUID, version int, createdAt time.Time, updatedAt time.Time, deletedAt *time.Time, issuedBy User) *AggregateBase[T] {
	return &AggregateBase[T]{
		aggregateId:      aggregateId,
		aggregateVersion: version,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
		issuedBy:         issuedBy,
		events:           make([]Event[T], 0),
	}
}

func (a AggregateBase[T]) AggregateId() uuid.UUID {
	return a.aggregateId
}

// Init is used to initialize an aggregate from an event
func (a *AggregateBase[T]) Init(e Event[T]) {
	a.aggregateId = e.AggregateId()
	a.createdAt = e.IssuedAt()
	a.issuedBy = e.IssuedBy()
	a.Process(e)
}

// Delete is used to mark an aggregate as deleted from an event
func (a *AggregateBase[T]) Delete(e Event[T]) {
	now := e.IssuedAt()
	a.deletedAt = &now
	a.Process(e)
}

// Process is used to track processing of an event
func (a *AggregateBase[T]) Process(e Event[T]) {
	// TODO we could increment the version here and check that it matches the event version
	a.aggregateVersion = e.AggregateVersion()
	a.events = append(a.events, e)
	a.updatedAt = e.IssuedAt()
	log.Debug().
		Str("aggregate_id", e.AggregateId().String()).
		Int("aggregate_version", a.AggregateVersion()).
		Str("aggregate_type", string(e.AggregateType())).
		Str("event_type", string(e.EventType())).
		Str("event_id", string(e.Id().String())).
		Msg("processing event")
}

func (a AggregateBase[T]) AggregateVersion() int {
	return a.aggregateVersion
}

func (a AggregateBase[T]) Events() []Event[T] {
	return a.events
}

func (a AggregateBase[T]) IssuedBy() User {
	return a.issuedBy
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
