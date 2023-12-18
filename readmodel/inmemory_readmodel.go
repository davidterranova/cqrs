package readmodel

import (
	"context"
	"fmt"
	"sync"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type InMemoryReadModel[T eventsourcing.Aggregate] struct {
	aggregates []*T
	sync.RWMutex

	*GenericHandler[T]
}

type AggregateMatcher[T eventsourcing.Aggregate] func(u *T) bool

func NewInMemoryReadModel[T eventsourcing.Aggregate](
	eventStream eventsourcing.Subscriber[T],
	aggregateFactory eventsourcing.AggregateFactory[T],
	evtTypeCreated eventsourcing.EventType,
	evtTypeDeleted eventsourcing.EventType,
) *InMemoryReadModel[T] {
	rm := &InMemoryReadModel[T]{
		aggregates: []*T{},
	}

	rm.GenericHandler = NewGenericHandler[T](
		aggregateFactory,
		evtTypeCreated,
		evtTypeDeleted,
		rm.create,
		rm.update,
		rm.delete,
		eventStream,
	)

	return rm
}

func (rM *InMemoryReadModel[T]) Find(_ context.Context, query AggregateMatcher[T]) ([]*T, error) {
	return rM.find(query), nil
}

func (rM *InMemoryReadModel[T]) Get(_ context.Context, query AggregateMatcher[T]) (*T, error) {
	matched := rM.find(query)
	if len(matched) == 0 {
		return nil, ErrNotFound
	}

	return matched[0], nil
}

func (rM *InMemoryReadModel[T]) create(aggregate *T) error {
	rM.RWMutex.Lock()
	defer rM.RWMutex.Unlock()

	rM.aggregates = append(rM.aggregates, aggregate)

	return nil
}

func (rM *InMemoryReadModel[T]) update(aggregateId uuid.UUID, fnRepo func(aggregate T) (T, error)) error {
	aggregates := rM.find(AggregateMatcherAggregateId[T](&aggregateId))
	if len(aggregates) == 0 {
		return ErrNotFound
	}

	updated, err := fnRepo(*aggregates[0])
	if err != nil {
		return fmt.Errorf("error updating aggregate: %w", err)
	}

	rM.RWMutex.Lock()
	defer rM.RWMutex.Unlock()
	*aggregates[0] = updated

	return nil
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

func (rM *InMemoryReadModel[T]) find(matcher AggregateMatcher[T]) []*T {
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
