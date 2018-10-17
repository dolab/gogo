package middlewares

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/gogo"
	"github.com/dolab/httptesting"
)

var (
	gogoapp  *gogo.AppServer
	gogotest *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		srcPath = path.Clean("../../")
	)

	gogoapp = gogo.New(runMode, srcPath)
	gogotest = httptesting.NewServer(gogoapp, false)

	code := m.Run()

	gogotest.Close()

	os.Exit(code)
}
