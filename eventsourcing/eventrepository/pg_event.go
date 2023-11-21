package eventrepository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidterranova/contacts/pkg/user"
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

func toPgEvent[T eventsourcing.Aggregate](e eventsourcing.Event[T]) (*pgEvent, error) {
	byteUser, err := e.IssuedBy().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal user", err)
	}

	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal event", err)
	}

	return &pgEvent{
		EventId:          e.Id(),
		EventType:        e.EventType(),
		EventIssuedAt:    e.IssuedAt(),
		EventIssuedBy:    string(byteUser),
		EventData:        data,
		AggregateId:      e.AggregateId(),
		AggregateType:    e.AggregateType(),
		AggregateVersion: e.AggregateVersion(),
	}, nil
}

func fromPgEvenSlice[T eventsourcing.Aggregate](registry eventsourcing.EventRegistry[T], pgEvents []pgEvent) ([]eventsourcing.Event[T], error) {
	events := make([]eventsourcing.Event[T], 0, len(pgEvents))
	for _, pgEvent := range pgEvents {
		hydratedEvent, err := fromPgEvent[T](registry, pgEvent)
		if err != nil {
			return nil, err
		}

		events = append(events, hydratedEvent)
	}

	return events, nil
}

func fromPgEvent[T eventsourcing.Aggregate](registry eventsourcing.EventRegistry[T], pgEvent pgEvent) (eventsourcing.Event[T], error) {
	u := user.New(uuid.Nil)
	err := json.Unmarshal([]byte(pgEvent.EventIssuedBy), &u)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal user", err)
	}

	return registry.Hydrate(
		*eventsourcing.NewEventBaseFromRepository[T](
			pgEvent.EventId,
			pgEvent.EventType,
			pgEvent.AggregateType,
			pgEvent.AggregateId,
			pgEvent.AggregateVersion,
			u,
		),
		pgEvent.EventData,
	)
}
