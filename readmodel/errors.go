package readmodel

import "errors"

var (
	ErrUnknownEvent = errors.New("unknown event")
	ErrNotFound     = errors.New("not found")
)
