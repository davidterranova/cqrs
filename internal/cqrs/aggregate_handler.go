package cqrs

import (
	"errors"
	"fmt"
)

type AggregateHandler interface {
	HandleCommand(cmd Command) (Aggregate, error)
	Load(aggregateID string) (Aggregate, error)
}

type AggregateFactory func() Aggregate

type aggregateHandler struct {
	eventStore       EventStore
	eventRegistry    EventRegistry
	aggregateFactory AggregateFactory
}

func NewAggregateHandler(eventStore EventStore, eventRegistry EventRegistry, aggregateFactory AggregateFactory) *aggregateHandler {
	return &aggregateHandler{
		eventStore:       eventStore,
		eventRegistry:    eventRegistry,
		aggregateFactory: aggregateFactory,
	}
}

func (h *aggregateHandler) HandleCommand(cmd Command) (Aggregate, error) {
	aggregate := h.aggregateFactory()

	if cmd.AggregateID() != EmptyAggregateID {
		events, err := h.eventStore.Load(cmd.AggregateID())
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("failed to load aggregate events: %w", err)
		}

		h.applyEventsToAggregate(aggregate, events)
	}

	events, err := cmd.Handle(aggregate)
	if err != nil {
		return nil, fmt.Errorf("command rejected: %w", err)
	}

	err = h.eventStore.Save(events...)
	if err != nil {
		return nil, fmt.Errorf("failed to persist events: %w", err)
	}

	h.applyEventsToAggregate(aggregate, events)

	return aggregate, nil
}

func (h *aggregateHandler) Load(aggregateID string) (Aggregate, error) {
	events, err := h.eventStore.Load(aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load aggregate events: %w", err)
	}

	aggregate := h.aggregateFactory()
	h.applyEventsToAggregate(aggregate, events)

	return aggregate, nil
}

func (h *aggregateHandler) applyEventsToAggregate(a Aggregate, evts []Event) {
	for _, event := range evts {
		if event.AggregateVersion() != a.Version() {
			panic("invalid aggregate version")
		}
		a.ApplyChangeHelper(a, event)
	}
}
