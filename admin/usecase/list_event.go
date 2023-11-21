package usecase

import (
	"context"

	"github.com/davidterranova/cqrs/eventsourcing"
)

type ListEventHandler[T eventsourcing.Aggregate] struct {
	repo eventsourcing.EventRepository[T]
}

func NewListEventHandler[T eventsourcing.Aggregate](repo eventsourcing.EventRepository[T]) *ListEventHandler[T] {
	return &ListEventHandler[T]{
		repo: repo,
	}
}

func (h *ListEventHandler[T]) Handle(ctx context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.Event[T], error) {
	return h.repo.Get(ctx, filter)
}
