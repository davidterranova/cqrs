package eventrepository

import (
	"context"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type inMemoryEventRepository struct {
	// events are stored by aggregate id
	aggregateEvents map[uuid.UUID][]inMemoryEvent

	// outbox for unpublished events
	outbox []inMemoryEvent
}

func NewInMemoryEventRepository() eventsourcing.EventRepository {
	return &inMemoryEventRepository{
		aggregateEvents: make(map[uuid.UUID][]inMemoryEvent),
		outbox:          make([]inMemoryEvent, 0),
	}
}

type inMemoryEvent struct {
	event     *eventsourcing.EventInternal
	published bool
}

func (r *inMemoryEventRepository) Save(_ context.Context, publishOutbox bool, events ...eventsourcing.EventInternal) error {
	for _, e := range events {
		e := e
		memoryEvent := inMemoryEvent{
			event:     &e,
			published: false,
		}

		aggregateId := e.AggregateId
		aggregateEvents, ok := r.aggregateEvents[aggregateId]
		if !ok {
			aggregateEvents = make([]inMemoryEvent, 0)
		}
		aggregateEvents = append(aggregateEvents, memoryEvent)
		r.aggregateEvents[aggregateId] = aggregateEvents

		// outbox
		r.outbox = append(r.outbox, memoryEvent)
	}

	return nil
}

func (r *inMemoryEventRepository) Get(_ context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.EventInternal, error) {
	events := make([]eventsourcing.EventInternal, 0)
	for _, me := range r.outbox {
		add := true

		if filter.AggregateId() != nil && *filter.AggregateId() != (*me.event).AggregateId {
			add = false
		}

		if filter.AggregateType() != nil && *filter.AggregateType() != (*me.event).AggregateType {
			add = false
		}

		if filter.EventType() != nil && *filter.EventType() != (*me.event).EventType {
			add = false
		}

		if filter.Published() != nil && *filter.Published() != me.published {
			add = false
		}

		if filter.IssuedBy() != nil && filter.IssuedBy().String() != (*me.event).EventIssuedBy {
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

		if filter.UpToVersion() != nil && (*me.event).AggregateVersion > *filter.UpToVersion() {
			add = false
		}

		if add {
			events = append(events, *me.event)
		}
	}

	return events, nil
}

func (r *inMemoryEventRepository) GetUnpublished(_ context.Context, batchSize int) ([]eventsourcing.EventInternal, error) {
	nbEvents := len(r.outbox)
	if batchSize < nbEvents {
		nbEvents = batchSize
	}

	unPublished := make([]eventsourcing.EventInternal, 0, nbEvents)
	for _, me := range r.outbox {
		if !me.published {
			unPublished = append(unPublished, *me.event)
		}
	}

	return unPublished, nil
}

func (r *inMemoryEventRepository) MarkAs(_ context.Context, asPublished bool, events ...eventsourcing.EventInternal) error {
	for _, e := range events {
		for _, me := range r.outbox {
			if (*me.event).AggregateId == e.AggregateId {
				me.published = asPublished
			}
		}
	}

	return nil
}
