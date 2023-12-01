package eventrepository

import (
	"context"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type inMemoryEventRepository struct {
	// events are stored by aggregate id
	aggregateEvents map[uuid.UUID][]*eventsourcing.EventInternal

	// outbox for unpublished events
	outbox []*eventsourcing.EventInternal
}

func NewInMemoryEventRepository() eventsourcing.EventRepository {
	return &inMemoryEventRepository{
		aggregateEvents: make(map[uuid.UUID][]*eventsourcing.EventInternal),
		outbox:          make([]*eventsourcing.EventInternal, 0),
	}
}

func (r *inMemoryEventRepository) Save(_ context.Context, publishOutbox bool, events ...eventsourcing.EventInternal) error {
	for _, e := range events {
		e := e

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

func (r *inMemoryEventRepository) Get(_ context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.EventInternal, error) {
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

func (r *inMemoryEventRepository) GetUnpublished(_ context.Context, batchSize int) ([]eventsourcing.EventInternal, error) {
	nbEvents := len(r.outbox)
	if batchSize < nbEvents {
		nbEvents = batchSize
	}

	unPublished := make([]eventsourcing.EventInternal, 0, nbEvents)
	for _, me := range r.outbox {
		if !me.EventPublished {
			unPublished = append(unPublished, *me)
		}
	}

	return unPublished, nil
}

func (r *inMemoryEventRepository) MarkAs(_ context.Context, asPublished bool, events ...eventsourcing.EventInternal) error {
	log.Info().Int("nb_events", len(events)).Bool("published", asPublished).Msg("marking events as")
	for _, e := range events {
		for _, me := range r.outbox {
			if me.EventId == e.EventId {
				log.Info().Str("event_id", me.EventId.String()).Str("event_type", string(me.EventType)).Bool("as published", asPublished).Msg("marking event")
				me.EventPublished = asPublished
			}
		}
	}

	return nil
}
