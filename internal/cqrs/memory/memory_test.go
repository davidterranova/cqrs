package memory

import (
	"testing"
	"time"

	"github.com/davidterranova/cqrs/internal/cqrs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type testEventData struct {
	ID        string
	CreatedAt time.Time
}

func (d testEventData) Route(e cqrs.Event, a cqrs.Aggregate) {}

func TestEventStore(t *testing.T) {
	var store cqrs.EventStore = NewEventStore()

	aggregate := cqrs.NewBaseAggregate()
	evtData := &testEventData{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}
	evt := cqrs.NewEvent().WithAggregate(aggregate).WithEventData(evtData)

	err := store.Save(evt)
	assert.NoError(t, err)

	evts, err := store.Load(aggregate.ID())
	assert.NoError(t, err)
	assert.Len(t, evts, 1)

	assert.Equal(t, evt, evts[0])
}
