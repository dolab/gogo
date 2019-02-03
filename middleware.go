package gogo

import "net/http"

// A MiddlewareConfiger defines config unmarshaler of middleware
type MiddlewareConfiger interface {
	Unmarshal(name string, v interface{}) error
}

// A Middlewarer defines a middleware of server
type Middlewarer interface {
	Name() string                                                                           // return name of middleware
	Config() []byte                                                                         // return config template required by middleware, in yaml format
	Priority() int                                                                          // sort priority?
	Register(MiddlewareConfiger) (func(w http.ResponseWriter, r *http.Request) bool, error) // return true to register middleware, false for do nothing
	Reload(MiddlewareConfiger) error                                                        // reload config at fly
	Shutdown() error                                                                        // for graceful quit
}

// A RequestReceivedMiddlewarer represents request received middleware interface of server
type RequestReceivedMiddlewarer interface {
	RequestReceived() []Middlewarer
}

// A RequestRoutedMiddlewarer represents request routed middleware interface of server
type RequestRoutedMiddlewarer interface {
	RequestRouted() []Middlewarer
}

// A ResponseReadyMiddlewarer represents response ready for sending data middleware interface of server
type ResponseReadyMiddlewarer interface {
	ResponseReady() []Middlewarer
}

// A ResponseAlwaysMiddlewarer represents response routed success middleware interface of server
type ResponseAlwaysMiddlewarer interface {
	ResponseAlways() []Middlewarer
}
