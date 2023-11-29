package eventsourcing

import "github.com/google/uuid"

func EnsureNewAggregate(aggregate Aggregate) error {
	if aggregate.AggregateId() != uuid.Nil || aggregate.AggregateVersion() != 0 {
		return ErrAggregateAlreadyExists
	}

	return nil
}

func EnsureAggregateNotNew(aggregate Aggregate) error {
	if aggregate.AggregateId() == uuid.Nil || aggregate.AggregateVersion() == 0 {
		return ErrAggregateNotFound
	}

	return nil
}
