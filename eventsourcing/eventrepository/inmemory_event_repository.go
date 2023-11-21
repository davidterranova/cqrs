package eventrepository

import (
	"context"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type inMemoryEventRepository[T eventsourcing.Aggregate] struct {
	// events are stored by aggregate id
	aggregateEvents map[uuid.UUID][]inMemoryEvent[T]

	// outbox for unpublished events
	outbox []inMemoryEvent[T]
}

type inMemoryEvent[T eventsourcing.Aggregate] struct {
	event     *eventsourcing.Event[T]
	published bool
}

func (r *inMemoryEventRepository[T]) Save(_ context.Context, publishOutbox bool, events ...eventsourcing.Event[T]) error {
	for _, e := range events {
		memoryEvent := inMemoryEvent[T]{
			event:     &e,
			published: false,
		}

		aggregateId := e.AggregateId()
		aggregateEvents, ok := r.aggregateEvents[aggregateId]
		if !ok {
			aggregateEvents = make([]inMemoryEvent[T], 0)
		}
		aggregateEvents = append(aggregateEvents, memoryEvent)
		r.aggregateEvents[aggregateId] = aggregateEvents

		// outbox
		if publishOutbox {
			r.outbox = append(r.outbox, memoryEvent)
		}
	}

	return nil
}

func (r *inMemoryEventRepository[T]) Get(_ context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.Event[T], error) {
	events := make([]eventsourcing.Event[T], 0)
	for _, me := range r.outbox {
		add := true

		if filter.AggregateId() != nil && *filter.AggregateId() != (*me.event).AggregateId() {
			add = false
		}

		if filter.AggregateType() != nil && *filter.AggregateType() != (*me.event).AggregateType() {
			add = false
		}

		if filter.EventType() != nil && *filter.EventType() != (*me.event).EventType() {
			add = false
		}

		if filter.Published() != nil && *filter.Published() != me.published {
			add = false
		}

		if filter.IssuedBy() != nil && filter.IssuedBy() != (*me.event).IssuedBy() {
			add = false
		}

		if filter.Limit() != nil && len(events) >= *filter.Limit() {
			add = false
		}

		// if filter.OrderBy() != nil {
		// 	// TODO
		// }

		// if filter.GroupBy() != nil {
		// 	// TODO
		// }

		if filter.UpToVersion() != nil && (*me.event).AggregateVersion() > *filter.UpToVersion() {
			add = false
		}

		if add {
			events = append(events, *me.event)
		}
	}

	return events, nil
}

func (r *inMemoryEventRepository[T]) GetUnpublished(_ context.Context, batchSize int) ([]eventsourcing.Event[T], error) {
	nbEvents := len(r.outbox)
	if batchSize < nbEvents {
		nbEvents = batchSize
	}

	unPublished := make([]eventsourcing.Event[T], 0, nbEvents)
	for _, me := range r.outbox {
		if !me.published {
			unPublished = append(unPublished, *me.event)
		}
	}

	return unPublished, nil
}

func (r *inMemoryEventRepository[T]) MarkAs(_ context.Context, asPublished bool, events ...eventsourcing.Event[T]) error {
	for _, e := range events {
		for _, me := range r.outbox {
			if (*me.event).AggregateId() == e.AggregateId() {
				me.published = asPublished
			}
		}
	}

	return nil
}
