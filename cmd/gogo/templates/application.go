package templates

var (
	applicationTemplate = `package controllers

import (
	"github.com/dolab/gogo"

	"{{.Namespace}}/{{.Application}}/app/middlewares"
	"{{.Namespace}}/{{.Application}}/app/models"
)

// An Application defines a service meeting gogo.Servicer interface, and wraps
// gogo.Grouper for custom resources.
type Application struct {
	v1 gogo.Grouper
}

// New creates a gogo.Servicer for intialization.
func New() gogo.Servicer {
	return &Application{}
}

// Init implements gogo.Servicer
func (app *Application) Init(config gogo.Configer, group gogo.Grouper) {
	err := NewAppConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// setup models
	err = models.Setup(Config.Model)
	if err != nil {
		panic(err.Error())
	}

	app.v1 = group.NewGroup("/v1")
}

// Filters implements gogo.Servicer
func (app *Application) Filters() {
	// apply your filters for group

	// panic recovery
	app.v1.Use(middlewares.Recovery())
}

// Resources implements gogo.Servicer
func (app *Application) Resources() {
	// register your resources
	// app.v1.GET("/", handler)

	app.v1.GET("/@greeting/hello", GettingStart.Hello)
}
`
)
