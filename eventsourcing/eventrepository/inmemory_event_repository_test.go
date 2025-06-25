//go:build unit

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

// mockUser implements the eventsourcing.User interface for testing
type mockUser struct {
	id   uuid.UUID
	name string
}

func (u *mockUser) Id() uuid.UUID             { return u.id }
func (u *mockUser) String() string            { return u.name }
func (u *mockUser) FromString(s string) error { u.name = s; return nil }

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

func TestInMemoryEventRepository_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryEventRepository()

	aggregateId := uuid.New()
	issuedBy := uuid.New().String()
	events := []eventsourcing.EventInternal{
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("created"),
			EventData:        []byte(`{"foo": "bar"}`),
			EventPublished:   false,
			AggregateId:      aggregateId,
			AggregateType:    "testType",
			AggregateVersion: 0,
		},
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("updated"),
			EventData:        []byte(`{"foo": "baz"}`),
			EventPublished:   true,
			AggregateId:      aggregateId,
			AggregateType:    "testType",
			AggregateVersion: 1,
		},
	}

	err := repo.Save(ctx, false, events...)
	assert.NoError(t, err)

	t.Run("Get by AggregateId", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateId(aggregateId),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("Get by AggregateType", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateType("testType"),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("Get by EventType", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithEventType("created"),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, eventsourcing.EventType("created"), res[0].EventType)
	})

	t.Run("Get by Published", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithPublished(true),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.True(t, res[0].EventPublished)
	})

	t.Run("Get by Unpublished", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithPublished(false),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.False(t, res[0].EventPublished)
	})

	t.Run("Get by IssuedBy", func(t *testing.T) {
		user := &mockUser{id: uuid.New(), name: issuedBy}
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithIssuedBy(user),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("Get with Limit", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithLimit(1),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Get with UpToVersion", func(t *testing.T) {
		res, err := repo.Get(ctx, eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithUpToVersion(0),
		))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 0, res[0].AggregateVersion)
	})
}

func TestInMemoryEventRepository_GetUnpublished(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryEventRepository()

	aggregateId := uuid.New()
	aggregateType := eventsourcing.AggregateType("testType")
	issuedBy := uuid.New().String()
	events := []eventsourcing.EventInternal{
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("created"),
			EventData:        []byte(`{"foo": "bar"}`),
			EventPublished:   false,
			AggregateId:      aggregateId,
			AggregateType:    aggregateType,
			AggregateVersion: 0,
		},
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("updated"),
			EventData:        []byte(`{"foo": "baz"}`),
			EventPublished:   true,
			AggregateId:      aggregateId,
			AggregateType:    aggregateType,
			AggregateVersion: 1,
		},
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("deleted"),
			EventData:        []byte(`{"foo": "qux"}`),
			EventPublished:   false,
			AggregateId:      aggregateId,
			AggregateType:    aggregateType,
			AggregateVersion: 2,
		},
	}

	err := repo.Save(ctx, false, events...)
	assert.NoError(t, err)

	t.Run("GetUnpublished returns only unpublished events and respects batch size", func(t *testing.T) {
		res, err := repo.GetUnpublished(ctx, aggregateType, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		for _, e := range res {
			assert.False(t, e.EventPublished)
			assert.Equal(t, aggregateType, e.AggregateType)
		}

		res, err = repo.GetUnpublished(ctx, aggregateType, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		for _, e := range res {
			assert.False(t, e.EventPublished)
			assert.Equal(t, aggregateType, e.AggregateType)
		}
	})
}
