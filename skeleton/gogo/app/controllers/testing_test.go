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
		srcPath = path.Clean("../../")
	)

	app := New(runMode, srcPath)
	app.Resources()

	gogotesting = httptesting.NewServer(app, false)

	code := m.Run()

	gogotesting.Close()

	os.Exit(code)
}
