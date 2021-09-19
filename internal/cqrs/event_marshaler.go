package cqrs

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventMarshaller interface {
	MarshalEvent(e Event) ([]byte, error)
	Unmarshal(data []byte) (Event, error)
}

type JSONEventMarshaller struct {
	eventRegistry EventRegistry
}

func NewJSONEventMarshaller(eventRegistry EventRegistry) *JSONEventMarshaller {
	return &JSONEventMarshaller{
		eventRegistry: eventRegistry,
	}
}

type jsonEvent struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	AggregateID   string `json:"aggregate_id"`
	AggregateType string `json:"aggregate_type"`

	EventType string `json:"event_type"`
	EventData string `json:"event_data"`
}

func (m JSONEventMarshaller) MarshalEvent(e Event) ([]byte, error) {
	encodedEventData, err := json.Marshal(e.EventData())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data %q: %w", e.EventType(), err)
	}

	internalEvent := jsonEvent{
		ID:        e.ID(),
		CreatedAt: e.CreatedAt(),

		AggregateID:   e.AggregateID(),
		AggregateType: e.AggregateType(),

		EventType: e.EventType(),
		EventData: string(encodedEventData),
	}

	return json.Marshal(internalEvent)
}

func (m JSONEventMarshaller) Unmarshal(data []byte) (Event, error) {
	var internalEvent jsonEvent

	err := json.Unmarshal(data, &internalEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	eventData, err := m.eventRegistry.NewEvent(internalEvent.EventType)
	if err != nil {
		return nil, fmt.Errorf("cannot create instance of event %q: %w", internalEvent.EventType, err)
	}

	err = json.Unmarshal([]byte(internalEvent.EventData), eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event %q: %w", internalEvent.EventType, err)
	}

	return &BaseEvent{
		id:        internalEvent.ID,
		createdAt: internalEvent.CreatedAt,

		aggregateID:   internalEvent.AggregateID,
		aggregateType: internalEvent.AggregateType,

		eventType: internalEvent.EventType,
		eventData: eventData,
	}, nil
}
