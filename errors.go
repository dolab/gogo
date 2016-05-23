package gogo

import (
	"errors"
)

var (
	ErrHeaderFlushed = errors.New("Response headers have been written!")
	ErrConfigSection = errors.New("Config section does not exist!")
	ErrSettingsKey   = errors.New("Settings key is duplicated!")
	ErrHash          = errors.New("The hash function does not linked into the binary!")
)
