package cqrs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const EmptyAggregateID = ""

type Aggregate interface {
	ID() string
	CreatedAt() time.Time
	Version() int

	String() string

	ApplyChangeHelper(a Aggregate, e Event)
}

type BaseAggregate struct {
	id        string
	createdAt time.Time
	version   int
}

func NewBaseAggregate() *BaseAggregate {
	return &BaseAggregate{
		id:        uuid.NewString(),
		createdAt: time.Now().UTC(),
	}
}

func (a *BaseAggregate) ID() string {
	return a.id
}

func (a *BaseAggregate) CreatedAt() time.Time {
	return a.createdAt
}

func (a *BaseAggregate) Version() int {
	return a.version
}

func (a *BaseAggregate) String() string {
	return fmt.Sprintf("id:%q created_at:%d", a.id, a.createdAt.Unix())
}

func (a *BaseAggregate) ApplyChangeHelper(ag Aggregate, e Event) {
	fmt.Printf("aggregate version:%d aggregateID:%q eventAggregateID:%q\n", ag.Version(), ag.ID(), e.AggregateID())
	if a.version == 0 {
		a.createdAt = e.CreatedAt()
		a.id = e.AggregateID()
	}

	e.EventData().Route(e, ag)
	a.version++
}
