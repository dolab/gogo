package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	newMockApp = func(mode string) *AppServer {
		root, _ := os.Getwd()

		return New(mode, path.Join(root, "skeleton", "gogo"))
	}
)

func Test_New(t *testing.T) {
	assertion := assert.New(t)

	app := newMockApp("test")
	app.GET("/gogo", func(ctx *Context) {
		ctx.Text("Hello, gogo!")
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	response, err := http.Get(ts.URL + "/gogo")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(response.Body)
	assertion.Nil(err)
	assertion.Equal("Hello, gogo!", string(body))
}

func Test_NewWithModeConfig(t *testing.T) {
	assertion := assert.New(t)

	app := newMockApp("development")
	assertion.Equal("for development", app.config.Name)
}
