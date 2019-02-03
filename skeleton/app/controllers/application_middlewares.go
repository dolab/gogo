package controllers

import (
	"net/http"

	"github.com/dolab/gogo"
)

type middleware struct {
	name string
	fn   func(w http.ResponseWriter, r *http.Request) bool
}

func (m *middleware) Name() string {
	return m.name
}

func (m *middleware) Config() []byte {
	return nil
}

func (m *middleware) Priority() int {
	return 0
}

func (m *middleware) Register(config gogo.MiddlewareConfiger) (func(w http.ResponseWriter, r *http.Request) bool, error) {
	return m.fn, nil
}

func (m *middleware) Reload(config gogo.MiddlewareConfiger) error {
	return nil
}

func (m *middleware) Shutdown() error {
	return nil
}

// RequestReceived allows custom middlewares for request received phase of server
func (app *Application) RequestReceived() []gogo.Middlewarer {
	return []gogo.Middlewarer{
		&middleware{
			name: "request_received@debugger",
			fn: func(w http.ResponseWriter, r *http.Request) bool {
				if Config.Debug {
					log := gogo.NewContextLogger(r)
					log.Debug("RequestReceivedHook")
				}

				return true
			},
		},
	}
}

// RequestRouted allows custom middlewares for request routed phase of server
func (app *Application) RequestRouted() []gogo.Middlewarer {
	return []gogo.Middlewarer{
		&middleware{
			name: "request_routed@debugger",
			fn: func(w http.ResponseWriter, r *http.Request) bool {
				if Config.Debug {
					log := gogo.NewContextLogger(r)
					log.Debug("RequestRoutedHook")
				}

				return true
			},
		},
	}
}

// ResponseReady allows custom middlewares for response ready phase of server
func (app *Application) ResponseReady() []gogo.Middlewarer {
	return []gogo.Middlewarer{
		&middleware{
			name: "response_ready@debugger",
			fn: func(w http.ResponseWriter, r *http.Request) bool {
				if Config.Debug {
					log := gogo.NewContextLogger(r)
					log.Debug("ResponseReadyHook")
				}

				return true
			},
		},
	}
}

// ResponseAlways allows custom middlewares for response always phase of server
func (app *Application) ResponseAlways() []gogo.Middlewarer {
	return []gogo.Middlewarer{
		&middleware{
			name: "response_always@debugger",
			fn: func(w http.ResponseWriter, r *http.Request) bool {
				if Config.Debug {
					log := gogo.NewContextLogger(r)
					log.Debug("ResponseAlwaysHook")
				}

				return true
			},
		},
	}
}
