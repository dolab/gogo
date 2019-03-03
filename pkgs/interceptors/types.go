package interceptors

import "net/http"

// A Interceptor is a callee for server
type Interceptor func(w http.ResponseWriter, r *http.Request) bool

// A Configer defines config unmarshaler for interceptors
type Configer interface {
	Unmarshal(name string, v interface{}) error
}

// A Interface defines a interface of server interceptor
type Interface interface {
	// Name returns name of interceptors
	Name() string

	// Config returns a template defined by interceptors, in yaml format
	Config() []byte

	// Priority returns sort priority of interceptors
	Priority() int

	// Register tries to register a interceptors with Configer.
	// The interceptor returned will be registered to server if error is nil.
	// Otherwise, nothing to do.
	Register(Configer) (Interceptor, error)

	// Realod tries to reload interceptors' config at fly.
	Reload(Configer) error
}

// A RequestReceivedInterceptor represents request received interface of server
type RequestReceivedInterceptor interface {
	RequestReceived() []Interface
}

// A RequestRoutedInterceptor represents request routed interface of server
type RequestRoutedInterceptor interface {
	RequestRouted() []Interface
}

// A ResponseReadyInterceptor represents response ready for sending data interface of server
type ResponseReadyInterceptor interface {
	ResponseReady() []Interface
}

// A ResponseAlwaysInterceptor represents response routed success interface of server
type ResponseAlwaysInterceptor interface {
	ResponseAlways() []Interface
}
