package controllers

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"
)

var (
	testClient *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	app := New(runMode, srcPath)
	app.Resources()

	testClient = httptesting.NewServer(app, false)

	code := m.Run()

	testClient.Close()

	os.Exit(code)
}
