package middleware

import (
	"errors"
)

// common errors
var (
	ErrInvalidPhase = errors.New("invalid phase of middleware, available values are RequestReceived, RequestRouted, ResponseReady and ResponseAlways")
)
