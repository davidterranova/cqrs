package eventrepository

import (
	"context"
	"sync"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type inMemoryEventRepository struct {
	// events are stored by aggregate id
	aggregateEvents map[uuid.UUID][]*eventsourcing.EventInternal
	// outbox for unpublished events
	outbox []*eventsourcing.EventInternal
	mtx    sync.RWMutex
}

func NewInMemoryEventRepository() eventsourcing.EventRepository {
	return &inMemoryEventRepository{
		aggregateEvents: make(map[uuid.UUID][]*eventsourcing.EventInternal),
		outbox:          make([]*eventsourcing.EventInternal, 0),
	}
}

func (r *inMemoryEventRepository) Save(_ context.Context, publishOutbox bool, events ...eventsourcing.EventInternal) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for _, e := range events {
		e := e
		log.Debug().
			Str("event_id", e.EventId.String()).
			Str("event_type", string(e.EventType)).
			Str("aggregate_type", string(e.AggregateType)).
			Str("aggregate_id", e.AggregateId.String()).
			Msg("event repository: saving event")

		aggregateId := e.AggregateId
		aggregateEvents, ok := r.aggregateEvents[aggregateId]
		if !ok {
			aggregateEvents = make([]*eventsourcing.EventInternal, 0)
		}
		aggregateEvents = append(aggregateEvents, &e)
		r.aggregateEvents[aggregateId] = aggregateEvents

		// outbox
		r.outbox = append(r.outbox, &e)
	}

	return nil
}

//nolint:cyclop
func (r *inMemoryEventRepository) Get(_ context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.EventInternal, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	events := make([]eventsourcing.EventInternal, 0)
	for _, me := range r.outbox {
		add := true

		if filter.AggregateId() != nil && *filter.AggregateId() != me.AggregateId {
			add = false
		}

		if filter.AggregateType() != nil && *filter.AggregateType() != me.AggregateType {
			add = false
		}

		if filter.EventType() != nil && *filter.EventType() != me.EventType {
			add = false
		}

		if filter.Published() != nil && *filter.Published() != me.EventPublished {
			add = false
		}

		if filter.IssuedBy() != nil && filter.IssuedBy().String() != me.EventIssuedBy {
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

		if filter.UpToVersion() != nil && me.AggregateVersion > *filter.UpToVersion() {
			add = false
		}

		if add {
			events = append(events, *me)
		}
	}

	return events, nil
}

func (r *inMemoryEventRepository) GetUnpublished(_ context.Context, aggregateType eventsourcing.AggregateType, batchSize int) ([]eventsourcing.EventInternal, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	nbEvents := len(r.outbox)
	if batchSize < nbEvents {
		nbEvents = batchSize
	}

	unPublished := make([]eventsourcing.EventInternal, 0, nbEvents)
	for _, me := range r.outbox {
		if !me.EventPublished && me.AggregateType == aggregateType {
			log.Debug().
				Str("event_type", string(me.EventType)).
				Str("event_id", me.EventId.String()).
				Str("aggregate_type", string(me.AggregateType)).
				Str("aggregate_id", me.AggregateId.String()).
				Msg("event repository: loading unpublished event")
			unPublished = append(unPublished, *me)
		}
	}

	return unPublished, nil
}

func (r *inMemoryEventRepository) MarkAs(_ context.Context, asPublished bool, events ...eventsourcing.EventInternal) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	log.Debug().
		Int("nb_events", len(events)).
		Bool("published", asPublished).
		Msg("marking events as")
	for _, e := range events {
		for _, me := range r.outbox {
			if me.EventId == e.EventId {
				log.Debug().
					Str("event_id", me.EventId.String()).
					Str("event_type", string(me.EventType)).
					Str("aggregate_type", string(me.AggregateType)).
					Bool("published", asPublished).
					Msg("event repository: marking event as")
				me.EventPublished = asPublished
			}
		}
	}

	return nil
}
