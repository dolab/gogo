package gogo

import (
	"github.com/dolab/gogo/pkgs/middleware"
)

// A RequestReceivedMiddlewarer represents request received middleware interface of server
type RequestReceivedMiddlewarer interface {
	RequestReceived() []middleware.Interface
}

// A RequestRoutedMiddlewarer represents request routed middleware interface of server
type RequestRoutedMiddlewarer interface {
	RequestRouted() []middleware.Interface
}

// A ResponseReadyMiddlewarer represents response ready for sending data middleware interface of server
type ResponseReadyMiddlewarer interface {
	ResponseReady() []middleware.Interface
}

// A ResponseAlwaysMiddlewarer represents response routed success middleware interface of server
type ResponseAlwaysMiddlewarer interface {
	ResponseAlways() []middleware.Interface
}
