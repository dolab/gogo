package gogo

import (
	"net/http"
	"net/http/httputil"

	"github.com/dolab/gogo/pkgs/hooks"
	"github.com/dolab/httpdispatch"
)

// A Configer represents config interface
type Configer interface {
	RunMode() RunMode
	RunName() string
	SetMode(mode RunMode)
	Section() *SectionConfig
	UnmarshalJSON(v interface{}) error
}

// A Grouper represents router interface
type Grouper interface {
	NewGroup(prefix string, middlewares ...Middleware) Grouper
	SetHandler(handler Handler)
	Use(middlewares ...Middleware)
	Resource(uri string, resource interface{}) Grouper
	OPTIONS(uri string, action Middleware)
	HEAD(uri string, action Middleware)
	POST(uri string, action Middleware)
	GET(uri string, action Middleware)
	PUT(uri string, action Middleware)
	PATCH(uri string, action Middleware)
	DELETE(uri string, action Middleware)
	Any(uri string, action Middleware)
	Static(uri, root string)
	Proxy(method, uri string, proxy *httputil.ReverseProxy)
	HandlerFunc(method, uri string, fn http.HandlerFunc)
	Handler(method, uri string, handler http.Handler)
	Handle(method, uri string, action Middleware)
	MockHandle(method, uri string, recorder http.ResponseWriter, action Middleware)
}

// A Servicer represents application interface
type Servicer interface {
	Init(config Configer, group Grouper)
	Middlewares()
	Resources()
}

// A Handler represents handler interface
type Handler interface {
	http.Handler

	Handle(string, string, httpdispatch.Handler)
	ServeFiles(string, http.FileSystem)
}

// A Middleware represents request filters or resource handler
//
// NOTE: It is the filter's responsibility to invoke ctx.Next() for chainning.
type Middleware func(ctx *Context)

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

// A RequestReceivedHooker represents request received hook interface of server
type RequestReceivedHooker interface {
	RequestReceivedHooks() []hooks.NamedHook
}

// A RequestRoutedHooker represents request routed hook interface of server
type RequestRoutedHooker interface {
	RequestRoutedHooks() []hooks.NamedHook
}

// A ResponseReadyHooker represents response ready for sending data hook interface of server
type ResponseReadyHooker interface {
	ResponseReadyHooks() []hooks.NamedHook
}

// A ResponseAlwaysHooker represents response routed success hook interface of server
type ResponseAlwaysHooker interface {
	ResponseAlwaysHooks() []hooks.NamedHook
}
