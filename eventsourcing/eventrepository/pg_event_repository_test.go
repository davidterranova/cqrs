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

// func TestEventStore(t *testing.T) {
// 	registry := eventsourcing.NewRegistry[domain.Contact]()
// 	registry.Register(domain.ContactCreated, func() eventsourcing.Event[domain.Contact] { return &domain.EvtContactCreated{} })
// 	registry.Register(domain.ContactNameUpdated, func() eventsourcing.Event[domain.Contact] { return &domain.EvtContactNameUpdated{} })
// 	registry.Register(domain.ContactEmailUpdated, func() eventsourcing.Event[domain.Contact] { return &domain.EvtContactEmailUpdated{} })
// 	registry.Register(domain.ContactPhoneUpdated, func() eventsourcing.Event[domain.Contact] { return &domain.EvtContactPhoneUpdated{} })
// 	registry.Register(domain.ContactDeleted, func() eventsourcing.Event[domain.Contact] { return &domain.EvtContactDeleted{} })

// 	db, err := pg.Open(pg.DBConfig{
// 		Name:       "postgres",
// 		ConnString: "postgres://postgres:password@127.0.0.1:5432/contacts?sslmode=disable&search_path=event_store",
// 	})
// 	require.NoError(t, err)
// 	store := eventsourcing.NewPGEventStore(db, registry)

// 	issuer := user.New(uuid.New())
// 	contactId := uuid.New()

// 	t.Run("Store", func(t *testing.T) {
// 		events := []eventsourcing.Event[domain.Contact]{
// 			domain.NewEvtContactCreated(contactId, issuer),
// 			domain.NewEvtContactNameUpdated(contactId, issuer, "John", "Doe"),
// 			domain.NewEvtContactEmailUpdated(contactId, issuer, "jdoe@contact.local"),
// 			domain.NewEvtContactPhoneUpdated(contactId, issuer, "+33612345678"),
// 		}

// 		err := store.Store(events...)
// 		require.NoError(t, err)
// 	})

// 	t.Run("Load", func(t *testing.T) {
// 		events, err := store.Load("contact", contactId)
// 		require.NoError(t, err)

// 		assert.NotEmpty(t, events)
// 	})
// }
