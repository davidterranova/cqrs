package eventsourcing

import (
	"context"
	"fmt"

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

type eventStore[T Aggregate] struct {
	repo       EventRepository
	registry   EventRegistry[T]
	withOutbox bool
}

func NewEventStore[T Aggregate](repo EventRepository, registry EventRegistry[T], withOutbox bool) *eventStore[T] {
	return &eventStore[T]{
		repo:       repo,
		registry:   registry,
		withOutbox: withOutbox,
	}
}

func (s *eventStore[T]) Store(ctx context.Context, events ...Event[T]) error {
	internalEvents, err := toEventInternalSlice[T](events)
	if err != nil {
		return fmt.Errorf("failed to convert events to internal events: %w", err)
	}

	return s.repo.Save(ctx, s.withOutbox, internalEvents...)
}

func (s *eventStore[T]) Load(ctx context.Context, aggregateType AggregateType, aggregateId uuid.UUID) ([]Event[T], error) {
	internalEvents, err := s.repo.Get(
		ctx,
		NewEventQuery(
			EventQueryWithAggregateType(aggregateType),
			EventQueryWithAggregateId(aggregateId),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load events from repository: %w", err)
	}

	events, err := FromEventInternalSlice[T](internalEvents, s.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to convert internal events to events: %w", err)
	}

	return events, nil
}

func (s *eventStore[T]) LoadUnpublished(ctx context.Context, batchSize int) ([]Event[T], error) {
	internalEvents, err := s.repo.GetUnpublished(ctx, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load unpublished events from repository: %w", err)
	}

	events, err := FromEventInternalSlice[T](internalEvents, s.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to convert internal events to events: %w", err)
	}

	return events, nil
}

func (s *eventStore[T]) MarkPublished(ctx context.Context, events ...Event[T]) error {
	internalEvents, err := toEventInternalSlice[T](events)
	if err != nil {
		return fmt.Errorf("failed to convert events to internal events: %w", err)
	}

	return s.repo.MarkAs(ctx, true, internalEvents...)
}

func (s *eventStore[T]) RepublishEvents(ctx context.Context, events ...Event[T]) error {
	internalEvents, err := toEventInternalSlice[T](events)
	if err != nil {
		return fmt.Errorf("failed to convert events to internal events: %w", err)
	}

	return s.repo.MarkAs(ctx, false, internalEvents...)
}
