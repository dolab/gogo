package middlewares

import (
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"

	"github.com/dolab/gogo/skeleton/app/controllers"
)

var (
	testApp    *controllers.Application
	testServer *httptest.Server
	testClient *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	testApp = controllers.New(runMode, srcPath)
	testServer = httptest.NewServer(testApp)
	testClient = httptesting.New(testServer.URL, false)

	code := m.Run()

	testServer.Close()

	os.Exit(code)
}
