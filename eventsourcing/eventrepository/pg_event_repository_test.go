//go:build integration
// +build integration

package eventrepository_test

import (
	"context"
	"testing"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/davidterranova/cqrs/eventsourcing/eventrepository"
	"github.com/davidterranova/cqrs/pg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGet(t *testing.T) {
	db := testDB(t)
	repo := eventrepository.NewPGEventRepository(db)
	ctx := context.Background()

	t.Run("it should load events", func(t *testing.T) {
		id := uuid.MustParse("cb310ab1-6284-4151-9d4d-d82428d548ea")

		filter := eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateId(id),
			eventsourcing.EventQueryWithAggregateType("contact"),
		)

		events, err := repo.Get(ctx, filter)
		assert.NoError(t, err)
		assert.NotEmpty(t, events)
	})

	t.Run("it should load pgEventOutbox entries", func(t *testing.T) {
		id := uuid.MustParse("cb310ab1-6284-4151-9d4d-d82428d548ea")

		filter := eventsourcing.NewEventQuery(
			eventsourcing.EventQueryWithAggregateId(id),
			eventsourcing.EventQueryWithAggregateType("contact"),
		)

		events, err := repo.Get(ctx, filter)
		assert.NoError(t, err)
		assert.NotEmpty(t, events)

		for _, event := range events {
			assert.True(t, event.EventPublished)
		}
	})
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := pg.Open(pg.DBConfig{
		Name:       "postgres",
		ConnString: "postgres://postgres:password@localhost:5432/contacts?sslmode=disable&search_path=event_store",
	})
	if err != nil {
		t.Fatal(err)
	}

	return db
}
