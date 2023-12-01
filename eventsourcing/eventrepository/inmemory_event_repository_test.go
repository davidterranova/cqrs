package eventrepository

import (
	"context"
	"testing"
	"time"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkAs(t *testing.T) {
	ctx := context.Background()
	t.Run("published", func(t *testing.T) {
		repo := NewInMemoryEventRepository()

		aggregateId := uuid.New()
		issuedBy := uuid.New().String()
		internalEvents := []eventsourcing.EventInternal{
			{
				EventId:          uuid.New(),
				EventIssuedAt:    time.Now().UTC(),
				EventIssuedBy:    issuedBy,
				EventType:        eventsourcing.EventType("created"),
				EventData:        []byte(`{}`),
				EventPublished:   false,
				AggregateId:      aggregateId,
				AggregateType:    "test",
				AggregateVersion: 0,
			},
			{
				EventId:          uuid.New(),
				EventIssuedAt:    time.Now().UTC(),
				EventIssuedBy:    issuedBy,
				EventType:        eventsourcing.EventType("name-set"),
				EventData:        []byte(`{"Name": "john"}`),
				EventPublished:   false,
				AggregateId:      aggregateId,
				AggregateType:    "test",
				AggregateVersion: 1,
			},
		}

		err := repo.Save(ctx, false, internalEvents...)
		assert.NoError(t, err)

		err = repo.MarkAs(ctx, true, internalEvents...)
		assert.NoError(t, err)

		events, err := repo.Get(
			ctx,
			eventsourcing.NewEventQuery(
				eventsourcing.EventQueryWithAggregateId(aggregateId),
			),
		)
		require.NoError(t, err)
		assert.Len(t, events, len(internalEvents))
		for _, e := range events {
			assert.True(t, e.EventPublished)
		}
	})
}
