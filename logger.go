package gogo

import (
	"log"
	"path"

	"github.com/dolab/logger"
)

// AppLogger implements Logger interface
type AppLogger struct {
	*logger.Logger

	requestId string
}

func NewAppLogger(output, filename string) *AppLogger {
	switch output {
	case "stdout", "stderr", "null", "nil":
		// skip

	default:
		if output[0] != '/' {
			output = path.Join(output, filename+".log")
		}
	}

	l, err := logger.New(output)
	if err != nil {
		log.Printf("logger.New(%s): %v\n", output, err)

		return nil
	}

	logger := &AppLogger{l, ""}
	return logger
}

// New returns a new Logger with provided requestId which shared writer with current logger
func (l *AppLogger) New(requestId string) Logger {
	copied := *l
	copied.Logger = copied.Logger.New(requestId)
	copied.requestId = requestId

	return &copied
}

func (l *AppLogger) RequestId() string {
	return l.requestId
}
