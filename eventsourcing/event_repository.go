package eventsourcing

import (
	"context"

	"github.com/google/uuid"
)

const (
	Published   = true
	Unpublished = false
)

type EventRepository interface {
	Save(ctx context.Context, publishOutbox bool, events ...EventInternal) error
	Get(ctx context.Context, filter EventQuery) ([]EventInternal, error)

	// load events from outbox that have not been published yet
	GetUnpublished(ctx context.Context, aggregateType AggregateType, batchSize int) ([]EventInternal, error)
	// MarkAs marks events as published / unpublished
	MarkAs(ctx context.Context, asPublished bool, events ...EventInternal) error
}

type EventQuery interface {
	AggregateId() *uuid.UUID
	AggregateType() *AggregateType
	EventType() *EventType
	Published() *bool
	IssuedBy() User
	Limit() *int
	OrderBy() (*string, *string)
	GroupBy() *string
	UpToVersion() *int
}
