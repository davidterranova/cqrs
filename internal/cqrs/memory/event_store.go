package memory

import (
	"sync"

	"github.com/davidterranova/cqrs/internal/cqrs"
)

type EventStore struct {
	events map[string][]cqrs.Event

	mtx sync.RWMutex
}

func NewEventStore() *EventStore {
	return &EventStore{
		events: make(map[string][]cqrs.Event),
	}
}

func (s *EventStore) Save(evts ...cqrs.Event) error {
	for _, e := range evts {
		err := s.save(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *EventStore) save(e cqrs.Event) error {
	s.mtx.RLock()
	eventList, ok := s.events[e.AggregateID()]
	s.mtx.RUnlock()

	if !ok {
		eventList = make([]cqrs.Event, 0, 1)
	}
	eventList = append(eventList, e)

	s.mtx.Lock()
	s.events[e.AggregateID()] = eventList
	s.mtx.Unlock()

	return nil
}

func (s *EventStore) Load(aggregateID string) ([]cqrs.Event, error) {
	s.mtx.RLock()
	eventList, ok := s.events[aggregateID]
	s.mtx.RUnlock()

	if !ok {
		return nil, cqrs.ErrNotFound
	}

	return eventList, nil
}
