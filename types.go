package gogo

import (
	"net/http"

	"github.com/dolab/httpdispatch"
)

// Configer represents application configurations
type Configer interface {
	RunMode() RunMode
	RunName() string
	SetMode(mode RunMode)
	Section() *SectionConfig
	UnmarshalJSON(v interface{}) error
}

// Handler represents server handlers
type Handler interface {
	http.Handler

	Handle(string, string, httpdispatch.Handler)
	ServeFiles(string, http.FileSystem)
}

// Middleware represents request filters and resource handler
//
// NOTE: It is the filter's responsibility to invoke ctx.Next() for chainning.
type Middleware func(ctx *Context)

// ResponseFilter defines filter interface applied to response
type ResponseFilter func(w Responser, b []byte) []byte

// Responser represents HTTP response interface
type Responser interface {
	http.ResponseWriter
	http.Flusher

	Before(filter ResponseFilter) // register before filter
	HeaderFlushed() bool          // whether response header has been sent?
	FlushHeader()                 // send response header
	Size() int                    // return the size of response body
	Status() int                  // response status code
	Hijack(w http.ResponseWriter) // hijack response with new http.ResponseWriter
}

// StatusCoder represents HTTP response status code
// it is useful for custom response data with response status code
type StatusCoder interface {
	StatusCode() int
}

// Logger defines interface of application log apis.
type Logger interface {
	New(requestID string) Logger
	Reuse(l Logger)
	RequestID() string
	SetLevelByName(level string) error
	SetColor(color bool)

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}
