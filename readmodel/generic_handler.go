package readmodel

import (
	"fmt"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// GenericHandler is a generic read model handler
// It can be used as a base for a read model handler and should be replaced
// in favor of a more specific handler when possible to increase performances.
// Generic approach requires the read model to load the aggregate, apply the event
// and save it back to the read model.
// Specific approaches should only update needed fields in the read model
// resulting in a single update query instead of a load, some processing, and a save.
type GenericHandler[T eventsourcing.Aggregate] struct {
	aggFactory     eventsourcing.AggregateFactory[T]
	evtTypeCreated eventsourcing.EventType
	evtTypeDeleted eventsourcing.EventType
	createFn       FnCreateRMAggregate[T]
	updateFn       FnUpdateRMAggregate[T]
	deleteFn       FnDeleteRMAggregate[T]
}

// FnCreateRMAggregate is a function that delegates the creation a read model aggregate
type FnCreateRMAggregate[T eventsourcing.Aggregate] func(a *T) error

// FnUpdateRMAggregate is a function that delegates the update of a read model aggregate
type FnUpdateRMAggregate[T eventsourcing.Aggregate] func(id uuid.UUID, fnRepo func(a T) (T, error)) error

// FnDeleteRMAggregate is a function that delegates the deletion of a read model aggregate
type FnDeleteRMAggregate[T eventsourcing.Aggregate] func(id uuid.UUID) error

func NewGenericHandler[T eventsourcing.Aggregate](
	aggFactory eventsourcing.AggregateFactory[T],
	evtTypeCreated eventsourcing.EventType,
	evtTypeDeleted eventsourcing.EventType,
	createFn FnCreateRMAggregate[T],
	updateFn FnUpdateRMAggregate[T],
	deleteFn FnDeleteRMAggregate[T],
	eventStream eventsourcing.Subscriber[T],
) *GenericHandler[T] {
	gh := &GenericHandler[T]{
		aggFactory:     aggFactory,
		evtTypeCreated: evtTypeCreated,
		evtTypeDeleted: evtTypeDeleted,
		createFn:       createFn,
		updateFn:       updateFn,
		deleteFn:       deleteFn,
	}

	if eventStream != nil {
		eventStream.Subscribe(gh.HandleEvent)
	}

	return gh
}

func (rm GenericHandler[T]) HandleEvent(e eventsourcing.Event[T]) {
	var err error

	switch e.EventType() {
	case rm.evtTypeCreated:
		agg := rm.aggFactory()
		err = e.Apply(agg)
		if err != nil {
			err = fmt.Errorf("error applying event: %w", err)
			break
		}

		err = rm.createFn(agg)
	case rm.evtTypeDeleted:
		err = rm.deleteFn(e.AggregateId())
	default:
		err = rm.updateFn(e.AggregateId(), func(agg T) (T, error) {
			err := e.Apply(&agg)
			if err != nil {
				return agg, fmt.Errorf("error applying event: %w", err)
			}

			return agg, nil
		})
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("aggregate_id", e.AggregateId().String()).
			Str("aggregate_type", string(e.AggregateType())).
			Str("event_type", e.EventType().String()).
			Msg("read model: error handling event")
	}
}
