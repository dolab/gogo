package controllers

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"
)

var (
	gogotest *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	app := New(runMode, srcPath)
	app.Resources()

	gogotest = httptesting.NewServer(app, false)

	code := m.Run()

	gogotest.Close()

	os.Exit(code)
}
