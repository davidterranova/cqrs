package eventsourcing

import (
	"github.com/google/uuid"
)

type eventQuery struct {
	aggregateId    *uuid.UUID
	aggregateType  *AggregateType
	eventType      *EventType
	published      *bool
	issuedBy       User
	limit          *int
	orderBy        *string
	orderDirection *string
	group_by       *string
	upToVersion    *int
}

type orderDirection string

const (
	ASC  orderDirection = "ASC"
	DESC orderDirection = "DESC"
)

type EventQueryOption func(*eventQuery)

func NewEventQuery(opts ...EventQueryOption) *eventQuery {
	eq := &eventQuery{}

	for _, opt := range opts {
		opt(eq)
	}

	return eq
}

func (eq *eventQuery) AggregateId() *uuid.UUID {
	return eq.aggregateId
}

func (eq *eventQuery) AggregateType() *AggregateType {
	return eq.aggregateType
}

func (eq *eventQuery) EventType() *EventType {
	return eq.eventType
}

func (eq *eventQuery) Published() *bool {
	return eq.published
}

func (eq *eventQuery) IssuedBy() User {
	return eq.issuedBy
}

func (eq *eventQuery) Limit() *int {
	return eq.limit
}

func (eq *eventQuery) OrderBy() (*string, *string) {
	return eq.orderBy, eq.orderDirection
}

func (eq *eventQuery) GroupBy() *string {
	return eq.group_by
}

func (eq *eventQuery) UpToVersion() *int {
	return eq.upToVersion
}

func EventQueryWithAggregateId(aggregateId uuid.UUID) EventQueryOption {
	return func(eq *eventQuery) {
		// if aggregateId != nil {
		eq.aggregateId = &aggregateId
		// }
	}
}

func EventQueryWithAggregateType(aggregateType AggregateType) EventQueryOption {
	return func(eq *eventQuery) {
		// if aggregateType != nil {
		eq.aggregateType = &aggregateType
		// }
	}
}

func EventQueryWithEventType(eventType EventType) EventQueryOption {
	return func(eq *eventQuery) {
		// if eventType != nil {
		eq.eventType = &eventType
		// }
	}
}

func EventQueryWithPublished(published bool) EventQueryOption {
	return func(eq *eventQuery) {
		// if published != nil {
		eq.published = &published
		// }
	}
}

func EventQueryWithIssuedBy(issuedBy User) EventQueryOption {
	return func(eq *eventQuery) {
		// if issuedBy != nil {
		eq.issuedBy = issuedBy
		// }
	}
}

func EventQueryWithLimit(limit int) EventQueryOption {
	return func(eq *eventQuery) {
		// if limit != nil {
		eq.limit = &limit
		// }
	}
}

func EventQueryWithOrderBy(orderBy string, orderDirection string) EventQueryOption {
	return func(eq *eventQuery) {
		// if orderBy != nil {
		eq.orderBy = &orderBy
		// }

		// if orderDirection != nil {
		eq.orderDirection = &orderBy
		// }
	}
}

func EventQueryWithGroupBy(group_by string) EventQueryOption {
	return func(eq *eventQuery) {
		// if group_by != nil {
		eq.group_by = &group_by
		// }
	}
}

func EventQueryWithUpToVersion(upToVersion int) EventQueryOption {
	return func(eq *eventQuery) {
		// if upToVersion != nil {
		eq.upToVersion = &upToVersion
		// }
	}
}
