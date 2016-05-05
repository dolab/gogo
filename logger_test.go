package gogo

import (
	"testing"

	"github.com/golib/assert"
)

func Test_NewAppLogger(t *testing.T) {
	assertion := assert.New(t)

	logger := NewAppLogger("null", "")
	assertion.NotNil(logger)
	assertion.Empty(logger.RequestId())

	// new with request id
	newLogger := logger.New("di-tseuqer-x")
	assertion.Empty(logger.RequestId())
	assertion.Equal("di-tseuqer-x", newLogger.RequestId())
	assertion.NotEqual(logger, newLogger)
}
