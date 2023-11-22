package usecase

import (
	"context"

	"github.com/davidterranova/cqrs/eventsourcing"
)

type ListEventHandler struct {
	repo eventsourcing.EventRepository
}

func NewListEventHandler(repo eventsourcing.EventRepository) *ListEventHandler {
	return &ListEventHandler{
		repo: repo,
	}
}

func (h *ListEventHandler) Handle(ctx context.Context, filter eventsourcing.EventQuery) ([]eventsourcing.EventInternal, error) {
	return h.repo.Get(ctx, filter)
}
