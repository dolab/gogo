package gogo

import (
	"net/http"
	"net/http/httputil"

	"github.com/dolab/httpdispatch"
)

// A Configer represents config interface
type Configer interface {
	RunMode() RunMode
	RunName() string
	SetMode(mode RunMode)
	Section() *SectionConfig
	UnmarshalJSON(v interface{}) error

	// for middlewares
	Middleware() MiddlewareConfiger
	LoadMiddlewares() error
}

// A Grouper represents router interface
type Grouper interface {
	NewGroup(prefix string, filters ...FilterFunc) Grouper
	SetHandler(handler Handler)
	Use(filters ...FilterFunc)
	Resource(uri string, resource interface{}) Grouper
	OPTIONS(uri string, handler FilterFunc)
	HEAD(uri string, handler FilterFunc)
	POST(uri string, handler FilterFunc)
	GET(uri string, handler FilterFunc)
	PUT(uri string, handler FilterFunc)
	PATCH(uri string, handler FilterFunc)
	DELETE(uri string, handler FilterFunc)
	Any(uri string, handler FilterFunc)
	Static(uri, root string)
	Proxy(method, uri string, proxy *httputil.ReverseProxy)
	HandlerFunc(method, uri string, fn http.HandlerFunc)
	Handler(method, uri string, handler http.Handler)
	Handle(method, uri string, handler FilterFunc)
	MockHandle(method, uri string, recorder http.ResponseWriter, handler FilterFunc)
}

// A Servicer represents application interface
type Servicer interface {
	Init(config Configer, group Grouper)
	Filters()
	Resources()
}

// A Handler represents handler interface
type Handler interface {
	http.Handler

	Handle(string, string, httpdispatch.Handler)
	ServeFiles(string, http.FileSystem)
}

// A Responser represents HTTP response interface
type Responser interface {
	http.ResponseWriter
	http.Flusher

	HeaderFlushed() bool        // whether response header has been sent?
	FlushHeader()               // send response header only if it has not sent
	Status() int                // response status code
	Size() int                  // return the size of response body
	Hijack(http.ResponseWriter) // hijack response with new http.ResponseWriter
}

// A StatusCoder represents HTTP response status code interface.
// it is useful for custom response data with response status code
type StatusCoder interface {
	StatusCode() int
}

// A Logger represents log interface
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

// A FilterFunc represents request filters or resource handler
//
// NOTE: It is the filter's responsibility to invoke ctx.Next() for chainning.
type FilterFunc func(ctx *Context)
