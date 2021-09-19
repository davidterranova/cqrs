package cqrs

import "errors"

var (
	ErrEventAlreadyRegistered = errors.New("event already registered")
	ErrUnknownEvent           = errors.New("unknown event")

	ErrNotFound = errors.New("not found")

	ErrInvalidAggregateType = errors.New("invalid aggregate type")
)
