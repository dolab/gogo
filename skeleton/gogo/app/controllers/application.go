package controllers

import (
	"github.com/dolab/gogo"

	"github.com/skeleton/app/middlewares"
)

type Application struct {
	*gogo.AppServer
}

func New(runMode, srcPath string) *Application {
	appServer := gogo.New(runMode, srcPath)

	err := NewAppConfig(appServer.Config())
	if err != nil {
		panic(err.Error())
	}

	return &Application{appServer}
}

// Middlerwares implements gogo.Middlewarer
// NOTE: DO NOT change the method name, its required by gogo!
func (app *Application) Middlewares() {
	// apply your middlewares

	// panic recovery
	app.Use(middlewares.Recovery())
}

// Resources implements gogo.Resourcer
// NOTE: DO NOT change the method name, its required by gogo!
func (app *Application) Resources() {
	// register your resources
	// app.GET("/", handler)

	app.GET("/@getting_start/hello", GettingStart.Hello)
}

// Run runs application after registering middelwares and resources
func (app *Application) Run() {
	// register middlewares
	app.Middlewares()

	// register resources
	app.Resources()

	// run server
	app.AppServer.Run()
}
