package memory

import (
	"reflect"
	"sync"

	"github.com/davidterranova/cqrs/internal/cqrs"
	"github.com/davidterranova/cqrs/internal/utils"
)

type EventRegistry struct {
	reg map[string]reflect.Type

	mtx sync.RWMutex
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		reg: make(map[string]reflect.Type),
	}
}

func (r *EventRegistry) Register(d cqrs.EventData) error {
	dataType, dataName := utils.GetTypeName(d)
	r.mtx.RLock()
	_, ok := r.reg[dataName]
	r.mtx.RUnlock()
	if ok {
		return cqrs.ErrEventAlreadyRegistered
	}

	r.mtx.Lock()
	r.reg[dataName] = dataType
	r.mtx.Unlock()
	return nil
}

func (r *EventRegistry) NewEvent(eventType string) (cqrs.EventData, error) {
	r.mtx.RLock()
	dataType, ok := r.reg[eventType]
	r.mtx.RUnlock()
	if !ok {
		return nil, cqrs.ErrUnknownEvent
	}

	return reflect.New(dataType).Interface().(cqrs.EventData), nil
}
