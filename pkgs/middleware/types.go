package middleware

import "net/http"

// A Configer defines config unmarshaler of middleware
type Configer interface {
	Unmarshal(name string, v interface{}) error
}

// A Interface defines a middleware of server hook
type Interface interface {
	// Name returns name of middleware
	Name() string

	// Config returns a template defined by middleware, in yaml format
	Config() []byte

	// Priority returns sort priority of middleware
	Priority() int

	// Register tries to register a middleware with Configer.
	// The interceptor returned will be registered to server if error is nil.
	// Otherwise, nothing to do.
	Register(Configer) (Interceptor, error)

	// Realod tries to reload middleware's config at fly.
	Reload(Configer) error

	// Shutdown is invoked when server is ready to close graceful.
	Shutdown() error
}

// A Interceptor is a callee for server
type Interceptor func(w http.ResponseWriter, r *http.Request) bool
