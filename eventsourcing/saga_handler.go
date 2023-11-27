package eventsourcing

type SagaHandler[T Aggregate] interface {
	// matching EventStream / SubscribeFn
	HandleEvent(e Event[T])

	Publish(Event[T])
}

/*
1. load saga related events from store
2. apply events to saga
3. publish new events
----

- Saga should maintain which events have been processed
- Saga should process new events and mark them as processed
*/

type SagaBase[T Aggregate] struct {
	events []Event[T]
}
