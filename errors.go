package gogo

import (
	"errors"
)

// errors
var (
	ErrConfigSection      = errors.New("Config section does not exist")
	ErrSettingsKey        = errors.New("Settings key is duplicated")
	ErrHeaderFlushed      = errors.New("Response headers have been written")
	ErrTooManyMiddlewares = errors.New("Too many middlewares")
	ErrReservedRoute      = errors.New("Reserved prefix of routes")
)
