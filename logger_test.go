package gogo

import (
	"testing"

	"github.com/golib/assert"
)

func Test_NewAppLogger(t *testing.T) {
	assertion := assert.New(t)

	logger := NewAppLogger("null", "")
	assertion.NotNil(logger)
	assertion.Empty(logger.RequestID())

	// new with request id
	newLogger := logger.New("di-tseuqer-x")
	assertion.Empty(logger.RequestID())
	assertion.Equal("di-tseuqer-x", newLogger.RequestID())
	assertion.NotEqual(logger, newLogger)
}
