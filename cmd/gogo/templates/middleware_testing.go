package templates

var (
	middlewareTestingTemplate = `package middlewares

import (
	"os"
	"path"
	"testing"

	"github.com/dolab/gogo"
	"github.com/dolab/httptesting"
)

var (
	gogoapp     *gogo.AppServer
	gogotesting *httptesting.Client
)

func TestMain(m *testing.M) {
	var (
		runMode = "test"
		cfgPath = path.Clean("../../")
	)

	gogoapp = gogo.New(runMode, cfgPath)
	gogotesting = httptesting.NewServer(gogoapp, false)

	code := m.Run()

	gogotesting.Close()

	os.Exit(code)
}
`
)
