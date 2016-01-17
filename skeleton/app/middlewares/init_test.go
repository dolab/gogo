package middlewares

import (
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/dolab/gogo"
	"github.com/dolab/httptesting"
)

var (
	testApp    *gogo.AppServer
	testServer *httptest.Server
	testClient *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	testApp = gogo.New(runMode, srcPath)
	testServer = httptest.NewServer(testApp)
	testClient = httptesting.New(testServer.URL, false)

	code := m.Run()

	testServer.Close()

	os.Exit(code)
}
