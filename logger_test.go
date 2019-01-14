package gogo

import (
	"testing"

	"github.com/golib/assert"
)

func Test_NewAppLogger(t *testing.T) {
	it := assert.New(t)

	logger := NewAppLogger("null", "")
	if it.NotNil(logger) {
		it.Implements((*Logger)(nil), logger)
		it.Empty(logger.RequestID())
	}
}

func Test_Logger_New(t *testing.T) {
	it := assert.New(t)

	logger := NewAppLogger("null", "")

	// new with request id
	tmplog := logger.New("di-tseuqer-x")
	if it.NotNil(tmplog) {
		it.Implements((*Logger)(nil), tmplog)
		it.NotEqual(logger, tmplog)
		it.Equal("di-tseuqer-x", tmplog.RequestID())
	}

	it.Empty(logger.RequestID())
}

func Benchmark_Logger_New(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	logger := NewAppLogger("nil", "")

	for i := 0; i < b.N; i++ {
		logger.New("logger")
	}
}

func Benchmark_Logger_Reuse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	logger := NewAppLogger("nil", "")

	for i := 0; i < b.N; i++ {
		tmplog := logger.New("logger")
		logger.Reuse(tmplog)
	}
}
