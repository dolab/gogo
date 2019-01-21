package gogo

import (
	"log"
	"net/http"
	"path"
	"sync"

	"github.com/dolab/logger"
)

// NewContextLogger returns a Logger related with *http.Request.
//
// NOTE: It returns a dummy *AppLogger when no available Logger for the request.
func NewContextLogger(r *http.Request) Logger {
	alog, ok := r.Context().Value(ctxLoggerKey).(Logger)
	if !ok {
		alog = NewAppLogger("stderr", "")
	}

	return alog
}

// AppLogger defines log component of gogo, it implements Logger interface
// with pool support
type AppLogger struct {
	*logger.Logger

	pool      sync.Pool
	requestID string
}

// NewAppLogger returns *AppLogger inited with args
func NewAppLogger(output, filename string) *AppLogger {
	switch output {
	case "stdout", "stderr", "null", "nil":
		// skip
	default:
		if output[0] != '/' {
			output = path.Join(output, filename+".log")
		}
	}

	lg, err := logger.New(output)
	if err != nil {
		log.Panicf("logger.New(%s): %v\n", output, err)
	}

	alog := &AppLogger{
		Logger: lg.New(),
	}

	// overwrite poo.New
	alog.pool.New = func() interface{} {
		return &AppLogger{
			Logger: lg.New(),
		}
	}

	return alog
}

// New returns a new Logger with provided requestID which shared writer with current logger
func (alog *AppLogger) New(requestID string) Logger {
	// shortcut
	if alog.requestID == requestID {
		return alog
	}

	lg := alog.pool.Get()
	if nlog, ok := lg.(*AppLogger); ok {
		nlog.requestID = requestID
		nlog.SetTags(requestID)

		return nlog
	}

	return lg.(Logger).New(requestID)
}

// RequestID returns request id binded to the logger
func (alog *AppLogger) RequestID() string {
	return alog.requestID
}

// Reuse puts the Logger back to pool for later usage
func (alog *AppLogger) Reuse(lg Logger) {
	alog.pool.Put(lg)
}
