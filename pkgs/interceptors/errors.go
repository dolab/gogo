package interceptors

import (
	"errors"
)

// common errors
var (
	ErrInvalidPhase = errors.New("invalid phase of interceptors, available values are RequestReceived, RequestRouted, ResponseReady and ResponseAlways")
)
