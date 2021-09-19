package cqrs

import (
	"fmt"
)

type Command interface {
	AggregateID() string
	Handle(a Aggregate) ([]Event, error)
}

type BaseCommand struct {
	aggregateID string
}

func NewBaseCommand(a Aggregate) *BaseCommand {
	var aggregateID string
	if a != nil {
		aggregateID = a.ID()
	}
	return &BaseCommand{
		aggregateID: aggregateID,
	}
}

func (c BaseCommand) AggregateID() string { return c.aggregateID }
func (c BaseCommand) String() string      { return fmt.Sprintf("aggregateID:%q", c.aggregateID) }
