package readmodel

import (
	"context"
	"testing"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const eventTypeNil eventsourcing.EventType = "event_type_nil"

type testAggregate struct {
	*eventsourcing.AggregateBase[testAggregate]
	value int
}

func (t testAggregate) AggregateType() eventsourcing.AggregateType {
	return "test_aggregate"
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

		return (*p).value == *value
	}
}

func TestInMemoryReadModelEventMatcher(t *testing.T) {
	ctx := context.Background()
	rm := NewInMemoryReadModel(nil, newTestAggregate, eventTypeNil, eventTypeNil, eventTypeNil)

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
