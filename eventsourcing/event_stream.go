package eventsourcing

import (
	"context"
)

type EventStream[T Aggregate] interface {
	Publisher[T]
	Subscriber[T]
}

type Publisher[T Aggregate] interface {
	Publish(ctx context.Context, events ...Event[T]) error
}

type SubscribeFn[T Aggregate] func(e Event[T])

type Subscriber[T Aggregate] interface {
	Subscribe(ctx context.Context, sub SubscribeFn[T])
}
