package eventstream

import (
	"context"
	"sync"

	"github.com/davidterranova/cqrs/eventsourcing"
	"github.com/rs/zerolog/log"
)

type eventStream[T eventsourcing.Aggregate] struct {
	stream      chan eventsourcing.Event[T]
	subscribers []eventsourcing.SubscribeFn[T]
	mtx         sync.RWMutex
	ctx         context.Context
}

func NewInMemoryPubSub[T eventsourcing.Aggregate](ctx context.Context, buffer int) *eventStream[T] {
	p := &eventStream[T]{
		ctx:         ctx,
		stream:      make(chan eventsourcing.Event[T], buffer),
		subscribers: make([]eventsourcing.SubscribeFn[T], 0),
	}
	go p.Run()

	return p
}

func (p *eventStream[T]) Publish(ctx context.Context, events ...eventsourcing.Event[T]) error {
	for _, event := range events {
		log.Ctx(ctx).Debug().Str("type", event.EventType().String()).Interface("event", event).Msg("publishing event")
		p.stream <- event
	}

	return nil
}

func (p *eventStream[T]) Run() {
	func() {
		for {
			select {
			case event := <-p.stream:
				p.mtx.RLock()
				for _, sub := range p.subscribers {
					sub(event)
				}
				p.mtx.RUnlock()
			case <-p.ctx.Done():
				close(p.stream)
				return
			}
		}
	}()
}

func (p *eventStream[T]) Subscribe(ctx context.Context, sub eventsourcing.SubscribeFn[T]) {
	p.mtx.Lock()
	p.subscribers = append(p.subscribers, sub)
	p.mtx.Unlock()
}
