package eventsourcing

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type EventStreamPublisher[T Aggregate] struct {
	eventRepo     EventRepository
	stream        Publisher[T]
	eventRegistry EventRegistry[T]
	batchSize     int
}

func NewEventStreamPublisher[T Aggregate](eventRepo EventRepository, eventRegistry EventRegistry[T], stream Publisher[T], batchSize int) *EventStreamPublisher[T] {
	return &EventStreamPublisher[T]{
		eventRepo:     eventRepo,
		eventRegistry: eventRegistry,
		stream:        stream,
		batchSize:     batchSize,
	}
}

func (p *EventStreamPublisher[T]) Run(ctx context.Context) {
	var (
		nb  int
		err error
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			nb, err = p.processBatch(ctx)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("event stream publisher: failed to process batch")
			}
		}

		var sleepTime time.Duration
		switch nb {
		case -1:
			sleepTime = 1 * time.Second
		case 0:
			sleepTime = 1 * time.Second
		default:
			sleepTime = 10 * time.Millisecond
		}
		time.Sleep(sleepTime)
	}
}

func (p *EventStreamPublisher[T]) processBatch(ctx context.Context) (int, error) {
	internalEvents, err := p.eventRepo.GetUnpublished(ctx, p.batchSize)
	if err != nil {
		return -1, err
	}

	if len(internalEvents) == 0 {
		return 0, nil
	}

	events, err := FromEventInternalSlice[T](internalEvents, p.eventRegistry)
	if err != nil {
		return -1, err
	}

	err = p.stream.Publish(ctx, events...)
	if err != nil {
		return -1, err
	}

	err = p.eventRepo.MarkAs(ctx, Published, internalEvents...)
	if err != nil {
		return -1, err
	}

	return len(events), nil
}
