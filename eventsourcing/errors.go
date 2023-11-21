package eventsourcing

import "errors"

var (
	ErrAggregateAlreadyExists = errors.New("aggregate already exists")
	ErrAggregateNotFound      = errors.New("aggregate not found")
	ErrInvalidAggregateType   = errors.New("invalid aggregate type")
	ErrUnknownEventType       = errors.New("unknown event type")
)
