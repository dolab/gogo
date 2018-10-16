package middlewares

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/gogo"
	"github.com/dolab/httptesting"
)

var (
	testApp    *gogo.AppServer
	testClient *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	testApp = gogo.New(runMode, srcPath)
	testClient = httptesting.NewServer(testApp, false)

	code := m.Run()

	testClient.Close()

	os.Exit(code)
}
