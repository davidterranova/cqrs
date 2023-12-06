//go:build unit

package readmodel

import (
	"context"
	"testing"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	eventTypeNil               eventsourcing.EventType     = "event_type_nil"
	testAggregateAggregateType eventsourcing.AggregateType = "test_aggregate"
)

type testAggregate struct {
	*eventsourcing.AggregateBase[testAggregate]
	value int
}

func (t testAggregate) AggregateType() eventsourcing.AggregateType {
	return testAggregateAggregateType
}

func newTestAggregate() *testAggregate {
	return &testAggregate{
		AggregateBase: eventsourcing.NewAggregateBase[testAggregate](uuid.Nil, 0),
	}
}

func newTestAggregateWithValue(value int) *testAggregate {
	return &testAggregate{
		AggregateBase: eventsourcing.NewAggregateBase[testAggregate](uuid.New(), 0),
		value:         value,
	}
}

func aggregateMatcherTestAggregateValue(value *int) AggregateMatcher[testAggregate] {
	return func(p *testAggregate) bool {
		if value == nil {
			return true
		}

		return p.value == *value
	}
}

//nolint:funlen
func TestInMemoryReadModelEventMatcher(t *testing.T) {
	ctx := context.Background()
	rm := NewInMemoryReadModel(nil, newTestAggregate, eventTypeNil, eventTypeNil)

	aggs := []*testAggregate{
		newTestAggregateWithValue(1),
		newTestAggregateWithValue(2),
		newTestAggregateWithValue(3),
		newTestAggregateWithValue(4),
		newTestAggregateWithValue(1),
	}
	rm.aggregates = append(rm.aggregates, aggs...)

	t.Run("match all", func(t *testing.T) {
		matchedAggregates, err := rm.Find(ctx, nil)
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, len(aggs))
	})

	t.Run("match by aggregateId", func(t *testing.T) {
		aggId := aggs[2].AggregateId()
		matchedAggregates, err := rm.Find(ctx, AggregateMatcherAggregateId[testAggregate](&aggId))
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 1)
		assert.Equal(t, aggs[2], matchedAggregates[0])
	})

	t.Run("match by value", func(t *testing.T) {
		one := 1
		matchedAggregates, err := rm.Find(ctx, aggregateMatcherTestAggregateValue(&one))
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 2)

		two := 2
		matchedAggregates, err = rm.Find(ctx, aggregateMatcherTestAggregateValue(&two))
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 1)
	})

	t.Run("match by aggregateId OR value", func(t *testing.T) {
		aggId := aggs[2].AggregateId()
		two := 2
		matchedAggregates, err := rm.Find(
			ctx,
			AggregateMatcherOr[testAggregate](
				aggregateMatcherTestAggregateValue(&two),
				AggregateMatcherAggregateId[testAggregate](&aggId),
			),
		)
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 2)
	})

	t.Run("match by aggregateId AND value", func(t *testing.T) {
		aggId := aggs[0].AggregateId()
		one := 1
		matchedAggregates, err := rm.Find(
			ctx,
			AggregateMatcherAnd[testAggregate](
				aggregateMatcherTestAggregateValue(&one),
				AggregateMatcherAggregateId[testAggregate](&aggId),
			),
		)
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 1)

		two := 2
		matchedAggregates, err = rm.Find(
			ctx,
			AggregateMatcherAnd[testAggregate](
				aggregateMatcherTestAggregateValue(&two),
				AggregateMatcherAggregateId[testAggregate](&aggId),
			),
		)
		require.NoError(t, err)
		assert.Len(t, matchedAggregates, 0)
	})
}

const (
	evtTypeTestAggregateCreated  eventsourcing.EventType = "testAggregate.created"
	evtTypeTestAggregateValueSet eventsourcing.EventType = "testAggregate.value-set"
)

type evtTestAggregateCreated struct {
	*eventsourcing.EventBase[testAggregate]
}

func newEvtTestAggregateCreated(aggregateId uuid.UUID, aggregateVersion int, issuedBy eventsourcing.User) *evtTestAggregateCreated {
	return &evtTestAggregateCreated{
		EventBase: eventsourcing.NewEventBase[testAggregate](
			testAggregateAggregateType,
			aggregateVersion,
			evtTypeTestAggregateCreated,
			aggregateId,
			issuedBy,
		),
	}
}

func (e evtTestAggregateCreated) Apply(a *testAggregate) error {
	a.Init(e)

	return nil
}

type evtTestAggregateValueSet struct {
	*eventsourcing.EventBase[testAggregate]
	value int
}

func newEvtTestAggregateValueSet(aggregateId uuid.UUID, aggregateVersion int, issuedBy eventsourcing.User, value int) *evtTestAggregateValueSet {
	return &evtTestAggregateValueSet{
		EventBase: eventsourcing.NewEventBase[testAggregate](
			testAggregateAggregateType,
			aggregateVersion,
			evtTypeTestAggregateValueSet,
			aggregateId,
			issuedBy,
		),
		value: value,
	}
}

func (e evtTestAggregateValueSet) Apply(a *testAggregate) error {
	a.Process(e)
	a.value = e.value

	return nil
}

func TestInMemoryReadModelHandleEvent(t *testing.T) {
	rm := NewInMemoryReadModel(nil, newTestAggregate, evtTypeTestAggregateCreated, eventTypeNil)

	// issuer := domain.NewUser()
	aggregateId1 := uuid.New()
	aggregateId2 := uuid.New()
	events := []eventsourcing.Event[testAggregate]{
		newEvtTestAggregateCreated(aggregateId1, 0, nil),
		newEvtTestAggregateValueSet(aggregateId1, 1, nil, 1),
		newEvtTestAggregateValueSet(aggregateId1, 2, nil, 2),
		newEvtTestAggregateCreated(aggregateId2, 0, nil),
		newEvtTestAggregateValueSet(aggregateId2, 1, nil, 5),
	}

	for _, e := range events {
		rm.HandleEvent(e)
	}

	t.Run("check number of aggregates", func(t *testing.T) {
		assert.Len(t, rm.aggregates, 2)
	})

	t.Run("check aggregate 1", func(t *testing.T) {
		agg, err := rm.Get(context.Background(), AggregateMatcherAggregateId[testAggregate](&aggregateId1))
		require.NoError(t, err)
		assert.Equal(t, aggregateId1, agg.AggregateId())
		assert.Equal(t, 2, agg.value)
	})

	t.Run("check aggregate 2", func(t *testing.T) {
		agg, err := rm.Get(context.Background(), AggregateMatcherAggregateId[testAggregate](&aggregateId2))
		require.NoError(t, err)
		assert.Equal(t, aggregateId2, agg.AggregateId())
		assert.Equal(t, 5, agg.value)
	})
}
