//go:build integration

package eventrepository_test

import (
	"context"
	"testing"
	"time"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/eventsourcing/eventrepository"
	"github.com/davidterranova/cqrs/pg"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPGEventRepository(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	err := db.Exec("TRUNCATE TABLE events CASCADE;").Error
	require.NoError(t, err)

	repo := eventrepository.NewPGEventRepository(db)
	log.Logger = log.Logger.Level(zerolog.InfoLevel)

	unpublishedAggregateId := uuid.New()
	issuedBy := uuid.New().String()
	unpublishedEvents := []eventsourcing.EventInternal{
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("created"),
			EventData:        []byte(`{}`),
			EventPublished:   false,
			AggregateId:      unpublishedAggregateId,
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
			AggregateId:      unpublishedAggregateId,
			AggregateType:    "test",
			AggregateVersion: 1,
		},
	}

	publishedAggregateId := uuid.New()
	publishedEvents := []eventsourcing.EventInternal{
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("created"),
			EventData:        []byte(`{}`),
			EventPublished:   true,
			AggregateId:      publishedAggregateId,
			AggregateType:    "test",
			AggregateVersion: 0,
		},
		{
			EventId:          uuid.New(),
			EventIssuedAt:    time.Now().UTC(),
			EventIssuedBy:    issuedBy,
			EventType:        eventsourcing.EventType("name-set"),
			EventData:        []byte(`{"Name": "Luke"}`),
			EventPublished:   true,
			AggregateId:      publishedAggregateId,
			AggregateType:    "test",
			AggregateVersion: 1,
		},
	}

	t.Run("save events with outbox", func(t *testing.T) {
		err := repo.Save(ctx, true, unpublishedEvents...)
		require.NoError(t, err)

		err = repo.Save(ctx, true, publishedEvents...)
		require.NoError(t, err)

		t.Run("find events by aggregateId", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithAggregateId(unpublishedAggregateId),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
		})

		t.Run("find events by aggregateType", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithAggregateType("test"),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 4)
		})

		t.Run("find events by eventType", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithEventType("name-set"),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
		})

		t.Run("find events by event published", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithPublished(true),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
		})

		t.Run("find events by event unpublished", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithPublished(false),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 2)
		})

		t.Run("find events by multiple criteria", func(t *testing.T) {
			events, err := repo.Get(
				ctx,
				eventsourcing.NewEventQuery(
					eventsourcing.EventQueryWithEventType("name-set"),
					eventsourcing.EventQueryWithAggregateId(publishedAggregateId),
				),
			)
			assert.NoError(t, err)
			assert.Len(t, events, 1)
		})

		t.Run("get unpublished events", func(t *testing.T) {
			events, err := repo.GetUnpublished(ctx, "test", 1)
			assert.NoError(t, err)
			assert.Len(t, events, 1)

			events, err = repo.GetUnpublished(ctx, "test", 10)
			assert.NoError(t, err)
			assert.Len(t, events, 2)

			t.Run("mark events as published", func(t *testing.T) {
				err = repo.MarkAs(ctx, eventsourcing.Published, events...)
				assert.NoError(t, err)

				events, err = repo.GetUnpublished(ctx, "test", 10)
				assert.NoError(t, err)
				assert.Len(t, events, 0)
			})
		})

	})
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	var dbCon pg.DBConfig
	err := envconfig.Process("POSTGRES", &dbCon)
	if err != nil {
		dbCon = pg.DBConfig{
			ConnString: "postgres://postgres:password@127.0.0.1:5435/cqrs?sslmode=disable&search_path=eventstore",
		}
	}

	db, err := pg.Open(dbCon)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
