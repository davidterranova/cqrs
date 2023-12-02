package eventsourcing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
)

type EventStreamPublisher[T Aggregate] struct {
	eventRepo     EventRepository
	stream        Publisher[T]
	eventRegistry EventRegistry[T]
	aggregateType AggregateType
	userFactory   UserFactory
	batchSize     int
	backoff       bool
}

func NewEventStreamPublisher[T Aggregate](eventRepo EventRepository, eventRegistry EventRegistry[T], aggregateType AggregateType, userFactory UserFactory, stream Publisher[T], batchSize int, backoff bool) *EventStreamPublisher[T] {
	return &EventStreamPublisher[T]{
		eventRepo:     eventRepo,
		eventRegistry: eventRegistry,
		aggregateType: aggregateType,
		userFactory:   userFactory,
		stream:        stream,
		batchSize:     batchSize,
		backoff:       backoff,
	}
}

func (p *EventStreamPublisher[T]) Run(ctx context.Context) {
	var b backoff.BackOff
	if !p.backoff {
		b = backoff.NewConstantBackOff(0 * time.Millisecond)
	} else {
		b = backoff.WithMaxRetries(
			backoff.WithContext(
				backoff.NewExponentialBackOff(),
				ctx,
			),
			5,
		)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = backoff.Retry(func() error {
				nb, err := p.processBatch(ctx)
				if err != nil {
					//! WARN: errors are not properly propagated and logged
					log.Ctx(ctx).Error().Err(err).Msg("event publisher: failed to process batch")
					return err
				}

				if nb == 0 {
					return errors.New("no events to publish")
				}

				return nil
			}, b)
		}
	}
}

func (p *EventStreamPublisher[T]) processBatch(ctx context.Context) (int, error) {
	internalEvents, err := p.eventRepo.GetUnpublished(ctx, p.aggregateType, p.batchSize)
	if err != nil {
		return -1, err
	}

	if len(internalEvents) == 0 {
		return 0, nil
	}

	events, err := FromEventInternalSlice[T](internalEvents, p.eventRegistry, p.userFactory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "event publisher: failed to convert internal events to events: (%p) %v\n", p.eventRegistry, err)
		return -1, fmt.Errorf("event publisher: failed to convert internal events to events: %w", err)
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
