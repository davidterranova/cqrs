package eventsourcing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidterranova/cqrs/user"
	"github.com/google/uuid"
)

type EventInternal struct {
	EventId          uuid.UUID
	EventIssuedAt    time.Time
	EventIssuedBy    user.User
	EventType        string
	EventData        []byte
	EventPublished   bool
	AggregateType    AggregateType
	AggregateId      uuid.UUID
	AggregateVersion int
}

func toEventInternalSlice[T Aggregate](events []Event[T]) ([]EventInternal, error) {
	internalEvents := make([]EventInternal, 0, len(events))
	for _, e := range events {
		internalEvent, err := toEventInternal(e)
		if err != nil {
			return nil, err
		}
		internalEvents = append(internalEvents, internalEvent)
	}

	return internalEvents, nil
}

func toEventInternal[T Aggregate](e Event[T]) (EventInternal, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return EventInternal{}, fmt.Errorf("%w: failed to marshal event", err)
	}

	return EventInternal{
		EventId:          e.Id(),
		EventIssuedAt:    e.IssuedAt(),
		EventIssuedBy:    e.IssuedBy(),
		EventType:        e.EventType(),
		EventData:        data,
		AggregateType:    e.AggregateType(),
		AggregateId:      e.AggregateId(),
		AggregateVersion: e.AggregateVersion(),
	}, nil
}

func FromEventInternalSlice[T Aggregate](internalEvents []EventInternal, registry EventRegistry[T]) ([]Event[T], error) {
	events := make([]Event[T], 0, len(internalEvents))
	for _, internalEvent := range internalEvents {
		event, err := fromEventInternal(internalEvent, registry)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func fromEventInternal[T Aggregate](internalEvent EventInternal, registry EventRegistry[T]) (Event[T], error) {
	return registry.Hydrate(
		*NewEventBaseFromRepository[T](
			internalEvent.EventId,
			internalEvent.EventType,
			internalEvent.EventIssuedBy,
			internalEvent.EventIssuedAt,
			internalEvent.AggregateType,
			internalEvent.AggregateId,
			internalEvent.AggregateVersion,
		),
		internalEvent.EventData,
	)
}
