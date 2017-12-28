package gogo

import (
	"log"
	"path"
	"sync"

	"github.com/dolab/logger"
)

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

	al := &AppLogger{
		Logger: lg,
	}
	al.pool.New = func() interface{} {
		return &AppLogger{
			Logger: lg.New(),
		}
	}

	return al
}

// New returns a new Logger with provided requestID which shared writer with current logger
func (al *AppLogger) New(requestID string) Logger {
	// shortcut
	if al.requestID == requestID {
		return al
	}

	nl := al.pool.Get().(*AppLogger)
	nl.SetTags(requestID)
	nl.requestID = requestID

	return nl
}

// RequestID returns request id binded to the logger
func (al *AppLogger) RequestID() string {
	return al.requestID
}

// Reuse puts the Logger back to pool for later usage
func (al *AppLogger) Reuse(lg Logger) {
	al.pool.Put(lg)
}
