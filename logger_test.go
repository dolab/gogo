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
	assertion.Implements((*Logger)(nil), logger)
}

func Test_Logger_New(t *testing.T) {
	assertion := assert.New(t)

	logger := NewAppLogger("null", "")

	// new with request id
	nlog := logger.New("di-tseuqer-x")
	assertion.NotNil(nlog)
	assertion.NotEqual(logger, nlog)
	assertion.Equal("di-tseuqer-x", nlog.RequestID())
	assertion.Implements((*Logger)(nil), nlog)

	assertion.Empty(logger.RequestID())
}

func Benchmark_Logger_New(b *testing.B) {
	logger := NewAppLogger("null", "")

	for i := 0; i < b.N; i++ {
		nlog := logger.New("logger")
		logger.Reuse(nlog)
	}
}
