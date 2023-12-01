package readmodel

import (
	"context"
	"sync"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AggregateMatcher is a function that returns true if the given aggregate matches the criteria
type AggregateMatcher[T eventsourcing.Aggregate] func(u *T) bool

type InMemoryReadModel[T eventsourcing.Aggregate] struct {
	aggregates []*T
	sync.RWMutex

	aggregateFactory eventsourcing.AggregateFactory[T]

	createdEventType  eventsourcing.EventType
	updatedEventTypes []eventsourcing.EventType
	deletedEventType  eventsourcing.EventType
}

func NewInMemoryReadModel[T eventsourcing.Aggregate](
	eventStream eventsourcing.Subscriber[T],
	aggregateFactory eventsourcing.AggregateFactory[T],
	createdEventType eventsourcing.EventType,
	deletedEventType eventsourcing.EventType,
	updatedEventTypes ...eventsourcing.EventType,
) *InMemoryReadModel[T] {
	rM := &InMemoryReadModel[T]{
		aggregates:        []*T{},
		aggregateFactory:  aggregateFactory,
		createdEventType:  createdEventType,
		updatedEventTypes: updatedEventTypes,
		deletedEventType:  deletedEventType,
	}

	if eventStream != nil {
		eventStream.Subscribe(context.Background(), rM.HandleEvent)
	}

	return rM
}

func (rM *InMemoryReadModel[T]) HandleEvent(e eventsourcing.Event[T]) {
	switch {
	case rM.isCreatedEvent(e):
		t := rM.aggregateFactory()
		err := e.Apply(t)
		if err != nil {
			log.Error().Err(err).Msgf("error applying event %s on %s %q", e.EventType(), e.AggregateType(), e.AggregateId())
			return
		}

		rM.RWMutex.Lock()
		rM.aggregates = append(rM.aggregates, t)
		rM.RWMutex.Unlock()
	case rM.isUpdatedEvent(e):
		aggregateId := e.AggregateId()
		t, err := rM.Get(context.Background(), AggregateMatcherAggregateId[T](&aggregateId))
		if err != nil {
			log.Error().Err(err).Msgf("error applying event %s on %s %q", e.EventType(), e.AggregateType(), e.AggregateId())
			return
		}

		err = e.Apply(t)
		if err != nil {
			log.Error().Err(err).Msgf("error applying event %s on %s %q", e.EventType(), e.AggregateType(), e.AggregateId())
			return
		}
	case rM.isDeletedEvent(e):
		err := rM.delete(e.AggregateId())
		if err != nil {
			log.Error().Err(err).Msgf("error applying event %s on %s %q", e.EventType(), e.AggregateType(), e.AggregateId())
			return
		}
	default:
		log.Error().Err(ErrUnknownEvent).Msgf("unknown event %s.%s", e.AggregateType(), e.EventType())
	}

	log.Debug().Str("event_id", e.Id().String()).Str("event_type", e.EventType().String()).Msg("read_model event applied")
}

func (rM *InMemoryReadModel[T]) isCreatedEvent(e eventsourcing.Event[T]) bool {
	return e.EventType() == rM.createdEventType
}

func (rM *InMemoryReadModel[T]) isUpdatedEvent(e eventsourcing.Event[T]) bool {
	for _, evtType := range rM.updatedEventTypes {
		if evtType == e.EventType() {
			return true
		}
	}

	return false
}

func (rM *InMemoryReadModel[T]) isDeletedEvent(e eventsourcing.Event[T]) bool {
	return e.EventType() == rM.deletedEventType
}

func (rM *InMemoryReadModel[T]) Find(_ context.Context, query AggregateMatcher[T]) ([]*T, error) {
	return rM.findAggregates(query), nil
}

func (rM *InMemoryReadModel[T]) Get(_ context.Context, query AggregateMatcher[T]) (*T, error) {
	matched := rM.findAggregates(query)
	if len(matched) == 0 {
		return nil, ErrNotFound
	}

	return matched[0], nil
}

func (rM *InMemoryReadModel[T]) delete(aggregateId uuid.UUID) error {
	for i, aggregate := range rM.aggregates {
		if (*aggregate).AggregateId() == aggregateId {
			rM.RWMutex.Lock()
			rM.aggregates = append(rM.aggregates[:i], rM.aggregates[i+1:]...)
			rM.RWMutex.Unlock()
			return nil
		}
	}

	return ErrNotFound
}

func (rM *InMemoryReadModel[T]) findAggregates(matcher AggregateMatcher[T]) []*T {
	var aggs []*T

	rM.RLock()
	defer rM.RUnlock()

	if matcher == nil {
		return rM.aggregates
	}

	for _, t := range rM.aggregates {
		if matcher(t) {
			aggs = append(aggs, t)
		}
	}

	return aggs
}

func AggregateMatcherAnd[T eventsourcing.Aggregate](matchers ...AggregateMatcher[T]) AggregateMatcher[T] {
	return func(p *T) bool {
		valid := true
		for _, m := range matchers {
			curr := m(p)
			valid = valid && curr
		}

		return valid
	}
}

func AggregateMatcherOr[T eventsourcing.Aggregate](matchers ...AggregateMatcher[T]) AggregateMatcher[T] {
	return func(p *T) bool {
		valid := false
		for _, m := range matchers {
			curr := m(p)
			valid = valid || curr
		}

		return valid
	}
}

func AggregateMatcherAggregateId[T eventsourcing.Aggregate](id *uuid.UUID) AggregateMatcher[T] {
	return func(p *T) bool {
		if id == nil {
			return true
		}

		return (*p).AggregateId() == *id
	}
}
