package usecase

import (
	"context"
	"fmt"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/user"
	"github.com/google/uuid"
)

const AllVersions = 0

type LoadAggregateHandler[T eventsourcing.Aggregate] struct {
	handler       eventsourcing.InternalCommandHandler[T]
	eventRepo     eventsourcing.EventRepository
	eventRegistry eventsourcing.EventRegistry[T]
	userFactory   user.UserFactory
	aggregateType eventsourcing.AggregateType
}

func NewLoadAggregateHandler[T eventsourcing.Aggregate](handler eventsourcing.InternalCommandHandler[T], eventRepo eventsourcing.EventRepository, eventRegistry eventsourcing.EventRegistry[T], userFactory user.UserFactory, aggregateType eventsourcing.AggregateType) *LoadAggregateHandler[T] {
	return &LoadAggregateHandler[T]{
		handler:       handler,
		eventRepo:     eventRepo,
		eventRegistry: eventRegistry,
		userFactory:   userFactory,
		aggregateType: aggregateType,
	}
}

func (h *LoadAggregateHandler[T]) Handle(ctx context.Context, aggregateId uuid.UUID, toVersion int) (*T, error) {
	if toVersion == AllVersions {
		return h.handler.HydrateAggregate(ctx, h.aggregateType, aggregateId)
	}

	internalEvents, err := h.eventRepo.Get(
		ctx,
		eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateId(aggregateId),
			eventsourcing.EventQueryWithAggregateType(h.aggregateType),
			eventsourcing.EventQueryWithUpToVersion(toVersion),
		),
	)
	if err != nil {
		return new(T), fmt.Errorf("failed to list events for aggregate(%s#%s): %w", h.aggregateType, aggregateId, err)
	}

	events, err := eventsourcing.FromEventInternalSlice[T](
		internalEvents,
		h.eventRegistry,
		h.userFactory,
	)
	if err != nil {
		return new(T), fmt.Errorf("failed to convert internal events to events for aggregate(%s#%s): %w", h.aggregateType, aggregateId, err)
	}

	return h.handler.HydrateAggregateFromEvents(ctx, h.aggregateType, events...)
}
