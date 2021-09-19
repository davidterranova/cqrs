package cqrs

type EventRegistry interface {
	Register(e EventData) error
	NewEvent(eventType string) (EventData, error)
}
