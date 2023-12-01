package eventsourcing

import (
	"encoding/json"
	"fmt"
)

type EventRegistry[T Aggregate] interface {
	Register(eventType EventType, factory func() Event[T])
	Hydrate(base EventBase[T], data []byte) (Event[T], error)
}

type eventRegistry[T Aggregate] struct {
	registry map[EventType]func() Event[T]
}

func NewEventRegistry[T Aggregate]() *eventRegistry[T] {
	return &eventRegistry[T]{
		registry: make(map[EventType]func() Event[T]),
	}
}

func (r *eventRegistry[T]) Register(eventType EventType, factory func() Event[T]) {
	r.registry[eventType] = factory
}

func (r eventRegistry[T]) create(eventType EventType) (Event[T], error) {
	factory, ok := r.registry[eventType]
	if !ok {
		return nil, fmt.Errorf("%w: event type %s not registered", ErrUnknownEventType, eventType)
	}

	return factory(), nil
}

func (r eventRegistry[T]) Hydrate(base EventBase[T], data []byte) (Event[T], error) {
	event, err := r.create(base.EventType())
	if err != nil {
		return nil, fmt.Errorf("failed to create empty event: %w", err)
	}

	err = json.Unmarshal(data, event)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	event.SetBase(base)

	return event, nil
}
