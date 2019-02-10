package debugger

import (
	"github.com/dolab/gogo/pkgs/middleware"
)

// A Registry defines two middlewares for server debugger
type Registry struct {
	requestReceived []middleware.Interface
	responseReady   []middleware.Interface
}

func NewRegistry() *Registry {
	return &Registry{
		requestReceived: []middleware.Interface{
			New(middleware.RequestReceived),
		},
		responseReady: []middleware.Interface{
			New(middleware.ResponseReady),
		},
	}
}

func (reg *Registry) RequestReceived() []middleware.Interface {
	return reg.requestReceived
}

func (reg *Registry) ResponseReady() []middleware.Interface {
	return reg.responseReady
}
