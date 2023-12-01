package http

import (
	"time"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type Event struct {
	EventId          uuid.UUID `json:"event_id"`
	EventType        string    `json:"event_type"`
	EventIssuedAt    time.Time `json:"event_issued_at"`
	EventIssuedBy    string    `json:"event_issued_by"`
	AggregateId      uuid.UUID `json:"aggregate_id"`
	AggregateType    string    `json:"aggregate_type"`
	AggregateVersion int       `json:"aggregate_version"`
	EventData        string    `json:"event_data"`
	EventPublished   bool      `json:"event_published"`
}

func fromEventInternalSlice(e []eventsourcing.EventInternal) []Event {
	events := make([]Event, len(e))
	for i, v := range e {
		events[i] = fromEventInternal(v)
	}
	return events
}

func fromEventInternal(e eventsourcing.EventInternal) Event {
	return Event{
		EventId:          e.EventId,
		EventType:        e.EventType.String(),
		EventIssuedAt:    e.EventIssuedAt,
		EventIssuedBy:    e.EventIssuedBy,
		EventPublished:   e.EventPublished,
		AggregateId:      e.AggregateId,
		AggregateType:    string(e.AggregateType),
		AggregateVersion: e.AggregateVersion,
		EventData:        string(e.EventData),
	}
}
