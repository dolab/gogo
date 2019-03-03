package debugger

import (
	"github.com/dolab/gogo/pkgs/interceptors"
)

// A Registry defines two interceptors for server debugger
type Registry struct {
	requestReceived []interceptors.Interface
	responseReady   []interceptors.Interface
}

func NewRegistry() *Registry {
	return &Registry{
		requestReceived: []interceptors.Interface{
			New(interceptors.RequestReceived),
		},
		responseReady: []interceptors.Interface{
			New(interceptors.ResponseReady),
		},
	}
}

func (reg *Registry) RequestReceived() []interceptors.Interface {
	return reg.requestReceived
}

func (reg *Registry) ResponseReady() []interceptors.Interface {
	return reg.responseReady
}
