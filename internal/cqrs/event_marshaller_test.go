package cqrs_test

import (
	"testing"
	"time"

	"github.com/davidterranova/cqrs/internal/cqrs"
	"github.com/davidterranova/cqrs/internal/cqrs/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEventData struct {
	ID        string
	CreatedAt time.Time
}

func (d testEventData) Route(e cqrs.Event, a cqrs.Aggregate) {}

func TestJSONEventMarshaller(t *testing.T) {
	r := memory.NewEventRegistry()
	err := r.Register(&testEventData{})
	require.NoError(t, err)

	m := cqrs.NewJSONEventMarshaller(r)

	eData := &testEventData{
		ID:        "edata_id_1",
		CreatedAt: time.Now().UTC(),
	}

	e := cqrs.NewEvent().WithAggregate(cqrs.NewBaseAggregate()).WithEventData(eData)

	jsonEvent, err := m.MarshalEvent(e)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonEvent)

	f, err := m.Unmarshal(jsonEvent)
	assert.NoError(t, err)

	assert.Equal(t, e, f, "expected to get \n%q \nbut got \n%q", e, f)
}
