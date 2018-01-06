package gogo

import (
	"fmt"
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

func Test_Logger_Reuse(t *testing.T) {
	assertion := assert.New(t)

	logger := NewAppLogger("null", "")

	nlog := logger.New("origin")
	logger.Reuse(nlog)

	tlog := logger.New("reuse")
	assertion.Equal(fmt.Sprintf("%p", nlog), fmt.Sprintf("%p", tlog))
}
