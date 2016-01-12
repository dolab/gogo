package gogo

import (
	"log"
	"path"

	"github.com/dolab/logger"
)

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
		log.Println("Cannot create logger:", err)

		return nil
	}

	return &AppLogger{l, ""}
}

func (l *AppLogger) New(requestId string) *AppLogger {
	copied := *l
	copied.Logger = copied.Logger.New(requestId)
	copied.requestId = requestId

	return &copied
}

func (l *AppLogger) SetRequestId(requestId string) *AppLogger {
	l.Logger = l.Logger.New(requestId)
	l.requestId = requestId

	return l
}

func (l *AppLogger) RequestId() string {
	return l.requestId
}
