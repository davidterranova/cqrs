package eventsourcing

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type EventStreamPublisher[T Aggregate] struct {
	eventRepo EventRepository[T]
	stream    Publisher[T]
	batchSize int
}

func NewEventStreamPublisher[T Aggregate](eventRepo EventRepository[T], stream Publisher[T], batchSize int) *EventStreamPublisher[T] {
	return &EventStreamPublisher[T]{
		eventRepo: eventRepo,
		stream:    stream,
		batchSize: batchSize,
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
	events, err := p.eventRepo.GetUnpublished(ctx, p.batchSize)
	if err != nil {
		return -1, err
	}

	if len(events) == 0 {
		return 0, nil
	}

	err = p.stream.Publish(ctx, events...)
	if err != nil {
		return -1, err
	}

	err = p.eventRepo.MarkAs(ctx, Published, events...)
	if err != nil {
		return -1, err
	}

	return len(events), nil
}
