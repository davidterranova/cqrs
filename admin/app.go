package admin

import (
	"context"

	"github.com/davidterranova/cqrs/admin/usecase"
	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const AllVersions = usecase.AllVersions

type App[T eventsourcing.Aggregate] struct {
	listEvent          *usecase.ListEventHandler
	loadAggregate      *usecase.LoadAggregateHandler[T]
	republishAggregate *usecase.RepublishAggregateHandler[T]
}

func NewwApp[T eventsourcing.Aggregate](listEvent *usecase.ListEventHandler, loadAggregate *usecase.LoadAggregateHandler[T], republishAggregate *usecase.RepublishAggregateHandler[T]) *App[T] {
	return &App[T]{
		listEvent:          listEvent,
		loadAggregate:      loadAggregate,
		republishAggregate: republishAggregate,
	}
}

func NewApp[T eventsourcing.Aggregate](eventRepository eventsourcing.EventRepository, registry eventsourcing.EventRegistry[T], aggregateType eventsourcing.AggregateType, factory eventsourcing.AggregateFactory[T]) *App[T] {
	// set to false to disable CQRS and remain in eventsourcing context
	CQRS := true
	eventstore := eventsourcing.NewEventStore[T](eventRepository, registry, CQRS)

	return &App[T]{
		listEvent: usecase.NewListEventHandler(eventRepository),
		loadAggregate: usecase.NewLoadAggregateHandler[T](
			eventsourcing.NewCommandHandler[T](eventstore, factory),
			eventRepository,
			registry,
			aggregateType,
		),
		republishAggregate: usecase.NewRepublishAggregateHandler[T](eventRepository), // should be set to nil if CQRS is disabled
	}
}

func (a *App[T]) ListEvent(ctx context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.EventInternal, error) {
	return a.listEvent.Handle(ctx, filter)
}

func (a *App[T]) LoadAggregate(ctx context.Context, aggregateId uuid.UUID, toVersion int) (*T, error) {
	return a.loadAggregate.Handle(ctx, aggregateId, toVersion)
}

func (a *App[T]) RepublishAggregate(ctx context.Context, aggregateId uuid.UUID) (int, error) {
	if a.republishAggregate == nil {
		log.Ctx(ctx).Warn().Msg("republishAggregate is nil, CQRS is disabled")
		return 0, nil
	}

	return a.republishAggregate.Handle(ctx, aggregateId)
}
