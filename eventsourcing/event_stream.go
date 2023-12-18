package eventsourcing

type EventStream[T Aggregate] interface {
	Publisher[T]
	Subscriber[T]
}

type Publisher[T Aggregate] interface {
	Publish(events ...Event[T]) error
}

type SubscribeFn[T Aggregate] func(e Event[T])

type Subscriber[T Aggregate] interface {
	Subscribe(sub SubscribeFn[T])
}
