package eventsourcing

import (
	"context"

	"github.com/davidterranova/cqrs/user"

	"github.com/google/uuid"
)

type EventStore[T Aggregate] interface {
	// Store events
	Store(ctx context.Context, events ...Event[T]) error
	// Load events from the given aggregate
	Load(ctx context.Context, aggregateType AggregateType, aggregateId uuid.UUID) ([]Event[T], error)
	// LoadUnpublished loads a batch of un published events
	LoadUnpublished(ctx context.Context, batchSize int) ([]Event[T], error)
	// MarkPublished marks events as published
	MarkPublished(ctx context.Context, events ...Event[T]) error
	// RepublishEvents republishes events so they can be consumed again
	RepublishEvents(ctx context.Context, events ...Event[T]) error
}

const (
	Published   = true
	Unpublished = false
)

type EventRepository[T Aggregate] interface {
	Save(ctx context.Context, publishOutbox bool, events ...Event[T]) error
	Get(ctx context.Context, filter EventQuery) ([]Event[T], error)

	// load events from outbox that have not been published yet
	GetUnpublished(ctx context.Context, batchSize int) ([]Event[T], error)
	// MarkAs marks events as published / unpublished
	MarkAs(ctx context.Context, asPublished bool, events ...Event[T]) error
}

type EventQuery interface {
	AggregateId() *uuid.UUID
	AggregateType() *AggregateType
	EventType() *string
	Published() *bool
	IssuedBy() user.User
	Limit() *int
	OrderBy() (*string, *string)
	GroupBy() *string
	UpToVersion() *int
}

type eventStore[T Aggregate] struct {
	repo       EventRepository[T]
	registry   EventRegistry[T]
	withOutbox bool
}

func NewEventStore[T Aggregate](repo EventRepository[T], registry EventRegistry[T], withOutbox bool) *eventStore[T] {
	return &eventStore[T]{
		repo:       repo,
		registry:   registry,
		withOutbox: withOutbox,
	}
}

func (s *eventStore[T]) Store(ctx context.Context, events ...Event[T]) error {
	return s.repo.Save(ctx, s.withOutbox, events...)
}

func (s *eventStore[T]) Load(ctx context.Context, aggregateType AggregateType, aggregateId uuid.UUID) ([]Event[T], error) {
	return s.repo.Get(
		ctx,
		NewEventQuery(
			EventQueryWithAggregateType(aggregateType),
			EventQueryWithAggregateId(aggregateId),
		),
	)
}

func (s *eventStore[T]) LoadUnpublished(ctx context.Context, batchSize int) ([]Event[T], error) {
	return s.repo.GetUnpublished(ctx, batchSize)
}

func (s *eventStore[T]) MarkPublished(ctx context.Context, events ...Event[T]) error {
	return s.repo.MarkAs(ctx, true, events...)
}

func (s *eventStore[T]) RepublishEvents(ctx context.Context, events ...Event[T]) error {
	return s.repo.MarkAs(ctx, false, events...)
}
