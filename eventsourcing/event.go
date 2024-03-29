package eventsourcing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const EvtTypeNil EventType = "-"

type Event[T Aggregate] interface {
	Id() uuid.UUID
	AggregateId() uuid.UUID
	AggregateType() AggregateType
	EventType() EventType
	IssuedAt() time.Time
	IssuedBy() User
	Apply(*T) error

	// SetBase(EventBase[T]) is used internally by eventsourcing package
	SetBase(EventBase[T])
	AggregateVersion() int
}

type EventType string

func (et EventType) String() string {
	return string(et)
}

type IEventRepository interface {
	Save(ctx context.Context, publishOutbox bool, events ...EventInternal) error
	Get(ctx context.Context, filter EventQuery) ([]EventInternal, error)

	// load events from outbox that have not been published yet
	GetUnpublished(ctx context.Context, batchSize int) ([]EventInternal, error)
	// MarkAs marks events as published / unpublished
	MarkAs(ctx context.Context, asPublished bool, events ...EventInternal) error
}
