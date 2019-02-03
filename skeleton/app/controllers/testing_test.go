package controllers

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"
)

var (
	gogotesting *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		cfgPath = path.Clean("../../")
	)

	app := gogo.New(runMode, cfgPath)
	app.NewResources(New())

	gogotesting = httptesting.NewServer(app, false)

	code := m.Run()

	gogotesting.Close()

	os.Exit(code)
}
