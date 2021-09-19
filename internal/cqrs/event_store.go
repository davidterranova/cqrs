package cqrs

type EventStore interface {
	Save(evts ...Event) error
	Load(aggregateID string) ([]Event, error)
}
