package eventrepository

import (
	"encoding/json"
	"time"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type pgEvent struct {
	EventId       uuid.UUID       `gorm:"type:uuid;primaryKey;column:event_id"`
	EventType     string          `gorm:"type:varchar(255);column:event_type"`
	EventIssuedAt time.Time       `gorm:"column:event_issued_at"`
	EventIssuedBy string          `gorm:"type:varchar(255);column:event_issued_by"`
	EventData     json.RawMessage `gorm:"type:jsonb;column:event_data"`

	AggregateId      uuid.UUID                   `gorm:"type:uuid;column:aggregate_id"`
	AggregateType    eventsourcing.AggregateType `gorm:"type:varchar(255);column:aggregate_type"`
	AggregateVersion int                         `gorm:"column:aggregate_version"`

	Outbox pgEventOutbox `gorm:"foreignKey:EventId;references:EventId"`
}

func (pgEvent) TableName() string {
	return "events"
}

type pgEventOutbox struct {
	EventId          uuid.UUID `gorm:"type:uuid;primaryKey;column:event_id"`
	Published        bool      `gorm:"column:published"`
	AggregateVersion int       `gorm:"column:aggregate_version"`
}

func (pgEventOutbox) TableName() string {
	return "events_outbox"
}

func toPgEvent(e eventsourcing.EventInternal) *pgEvent {
	return &pgEvent{
		EventId:          e.EventId,
		EventType:        e.EventType,
		EventIssuedAt:    e.EventIssuedAt,
		EventIssuedBy:    e.EventIssuedBy,
		EventData:        e.EventData,
		AggregateId:      e.AggregateId,
		AggregateType:    e.AggregateType,
		AggregateVersion: e.AggregateVersion,
	}
}

func fromPgEventSlice(pgEvents []pgEvent) ([]eventsourcing.EventInternal, error) {
	events := make([]eventsourcing.EventInternal, 0, len(pgEvents))
	for _, pgEvent := range pgEvents {
		events = append(events, fromPgEvent(pgEvent))
	}

	return events, nil
}

func fromPgEvent(pgEvent pgEvent) eventsourcing.EventInternal {
	return eventsourcing.EventInternal{
		EventId:          pgEvent.EventId,
		EventType:        pgEvent.EventType,
		EventIssuedAt:    pgEvent.EventIssuedAt,
		EventIssuedBy:    pgEvent.EventIssuedBy,
		EventData:        pgEvent.EventData,
		EventPublished:   pgEvent.Outbox.Published,
		AggregateId:      pgEvent.AggregateId,
		AggregateType:    pgEvent.AggregateType,
		AggregateVersion: pgEvent.AggregateVersion,
	}
}
