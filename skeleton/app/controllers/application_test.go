package controllers

import (
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"
)

var (
	testServer *httptest.Server
	testClient *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	app := New(runMode, srcPath)
	app.Resources()

	testServer = httptest.NewServer(app)
	testClient = httptesting.New(testServer.URL, false)

	code := m.Run()

	testServer.Close()

	os.Exit(code)
}
