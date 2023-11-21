package usecase

import (
	"context"
	"fmt"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
)

type RepublishAggregateHandler[T eventsourcing.Aggregate] struct {
	repo eventsourcing.EventRepository[T]
}

func NewRepublishAggregateHandler[T eventsourcing.Aggregate](repo eventsourcing.EventRepository[T]) *RepublishAggregateHandler[T] {
	return &RepublishAggregateHandler[T]{
		repo: repo,
	}
}

func (h *RepublishAggregateHandler[T]) Handle(ctx context.Context, aggregateId uuid.UUID) error {
	events, err := h.repo.Get(
		ctx,
		eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateId(aggregateId),
		),
	)
	if err != nil {
		return fmt.Errorf("republishAggregateHandler: failed to list aggregate events: %w", err)
	}

	err = h.repo.MarkAs(ctx, eventsourcing.Unpublished, events...)
	if err != nil {
		return fmt.Errorf("republishAggregateHandler: failed to republish aggregate events: %w", err)
	}

	return nil
}
