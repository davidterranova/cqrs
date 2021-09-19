package cqrs

import (
	"fmt"
	"time"

	"github.com/davidterranova/cqrs/internal/utils"
	"github.com/google/uuid"
)

type Event interface {
	ID() string
	CreatedAt() time.Time

	AggregateID() string
	AggregateType() string
	AggregateVersion() int

	EventType() string
	EventData() EventData

	String() string
}

type EventData interface {
	// TODO : implement Route(e Event, a Aggregate)
	Route(e Event, a Aggregate)
}

type BaseEvent struct {
	id        string
	createdAt time.Time

	aggregateID      string
	aggregateType    string
	aggregateVersion int

	eventType string
	eventData EventData
}

func NewEvent() *BaseEvent {
	return &BaseEvent{
		createdAt: time.Now().UTC(),
		id:        uuid.NewString(),
	}
}

func (b *BaseEvent) WithAggregate(a Aggregate) *BaseEvent {
	_, aggregateName := utils.GetTypeName(a)

	b.aggregateID = a.ID()
	b.aggregateType = aggregateName
	b.aggregateVersion = a.Version()

	return b
}

func (b *BaseEvent) WithEventData(d EventData) *BaseEvent {
	_, eventDataName := utils.GetTypeName(d)

	b.eventType = eventDataName
	b.eventData = d

	return b
}

func (b *BaseEvent) ID() string {
	return b.id
}

func (b *BaseEvent) CreatedAt() time.Time {
	return b.createdAt
}

func (b *BaseEvent) AggregateID() string {
	return b.aggregateID
}

func (b *BaseEvent) AggregateType() string {
	return b.aggregateType
}

func (b *BaseEvent) AggregateVersion() int {
	return b.aggregateVersion
}

func (b *BaseEvent) EventType() string {
	return b.eventType
}

func (b *BaseEvent) EventData() EventData {
	return b.eventData
}

func (b *BaseEvent) String() string {
	return fmt.Sprintf("id:%q created_at:%d aggregate_id:%q aggregate_type:%q event_type:%q event_data:%+v", b.id, b.createdAt.Unix(), b.aggregateID, b.aggregateType, b.eventType, b.eventData)
}
