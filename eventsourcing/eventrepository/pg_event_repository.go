package eventrepository

import (
	"context"
	"fmt"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/user"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type pgEventRepository[T eventsourcing.Aggregate] struct {
	db       *gorm.DB
	registry eventsourcing.EventRegistry[T]
}

func NewPGEventRepository[T eventsourcing.Aggregate](db *gorm.DB, registry eventsourcing.EventRegistry[T]) *pgEventRepository[T] {
	return &pgEventRepository[T]{
		db:       db,
		registry: registry,
	}
}

func (r pgEventRepository[T]) Save(ctx context.Context, publishOutbox bool, events ...eventsourcing.Event[T]) error {
	if len(events) == 0 {
		return nil
	}

	pgEvents := make([]*pgEvent, 0, len(events))
	outboxEntries := make([]*pgEventOutbox, 0, len(events))

	for _, event := range events {
		pgEvent, err := toPgEvent(event)
		if err != nil {
			return err
		}
		pgEvents = append(pgEvents, pgEvent)

		outboxEntries = append(outboxEntries, &pgEventOutbox{
			EventId:          event.Id(),
			Published:        false,
			AggregateVersion: event.AggregateVersion(),
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(pgEvents).Error
		if err != nil {
			return fmt.Errorf("failed to create events in event_store table: %w", err)
		}

		for _, event := range events {
			log.Debug().Str("type", event.EventType()).Interface("event", event).Msg("stored event")
		}

		if !publishOutbox {
			return nil
		}

		err = tx.Create(outboxEntries).Error
		if err != nil {
			return fmt.Errorf("failed to create events in event_outbox table: %w", err)
		}

		return nil
	})
}

func (r pgEventRepository[T]) Get(ctx context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.Event[T], error) {
	var pgEvents []pgEvent
	query := r.db.WithContext(ctx).
		Model(&pgEvent{}).
		Scopes(
			issuedByScope(filter.IssuedBy()),
			publishedScope(filter.Published()),
			eventTypeScope(filter.EventType()),
			aggregateTypeScope(filter.AggregateType()),
			aggregateIdScope(filter.AggregateId()),
			upToVersionScope(filter.UpToVersion()),
		).Joins("LEFT JOIN events_outbox ON events_outbox.event_id = events.event_id")

	if filter.Limit() != nil {
		query = query.Limit(*filter.Limit())
	}
	if filter.GroupBy() != nil {
		query = query.Group(*filter.GroupBy())
	}

	err := query.
		Debug().
		Find(&pgEvents).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get events from event_store table: %w", err)
	}

	events, err := fromPgEvenSlice[T](r.registry, pgEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate events from pgEvents: %w", err)
	}

	return events, nil
}

func (r pgEventRepository[T]) GetUnpublished(ctx context.Context, batchSize int) ([]eventsourcing.Event[T], error) {
	var pgOutboxEntries []uuid.UUID
	err := r.db.
		WithContext(ctx).
		Model(&pgEventOutbox{}).
		Where("published = ?", false).
		Group("event_id").
		Order("aggregate_version ASC").
		Limit(batchSize).
		Pluck("event_id", &pgOutboxEntries).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to load unpublished events from outbox: %w", err)
	}

	if len(pgOutboxEntries) == 0 {
		return nil, nil
	}

	var unpublishedEvents []pgEvent
	err = r.db.WithContext(ctx).Where("event_id IN ?", pgOutboxEntries).Find(&unpublishedEvents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load unpublished events: %w", err)
	}

	for _, event := range unpublishedEvents {
		log.Debug().Str("type", event.EventType).Interface("event", event).Msg("loaded unpublished event")
	}

	return fromPgEvenSlice(r.registry, unpublishedEvents)
}

func (r pgEventRepository[T]) MarkAs(ctx context.Context, asPublished bool, events ...eventsourcing.Event[T]) error {
	if len(events) == 0 {
		return nil
	}

	var eventIds []uuid.UUID
	for _, event := range events {
		eventIds = append(eventIds, event.Id())
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(&pgEventOutbox{}).Where("event_id IN ?", eventIds).Update("published", asPublished).Error
	})
}

func issuedByScope(user user.User) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user == nil {
			return db
		}

		return db.Where("events.event_issued_by = ?", user)
	}
}

func publishedScope(published *bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if published == nil {
			return db
		}

		return db.Where("events_outbox.published = ?", *published)
	}
}

func eventTypeScope(eventType *string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if eventType == nil {
			return db
		}

		return db.Where("events.event_type = ?", *eventType)
	}
}

func aggregateTypeScope(aggregateType *eventsourcing.AggregateType) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if aggregateType == nil {
			return db
		}

		return db.Where("events.aggregate_type = ?", *aggregateType)
	}
}

func aggregateIdScope(aggregateId *uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if aggregateId == nil {
			return db
		}

		return db.Where("events.aggregate_id = ?", *aggregateId)
	}
}

func upToVersionScope(version *int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if version == nil {
			return db
		}

		return db.Where("events_outbox.aggregate_version <= ?", *version)
	}
}
