package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golib/assert"
)

var (
	fakeApp = func(mode string) *AppServer {
		root, _ := os.Getwd()

		return New(mode, path.Join(root, "skeleton", "gogo"))
	}
)

func Test_New(t *testing.T) {
	assertion := assert.New(t)

	app := fakeApp("test")
	app.GET("/gogo", func(ctx *Context) {
		ctx.Text("Hello, gogo!")
	})
	app.PUT("/ping", func(ctx *Context) {
		ctx.Text("pong")
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	// GET
	response, err := http.Get(ts.URL + "/gogo")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	assertion.Nil(err)
	assertion.Equal("Hello, gogo!", string(body))

	// PUT
	request, _ := http.NewRequest(http.MethodPut, ts.URL+"/ping", nil)
	response, err = http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err = ioutil.ReadAll(response.Body)
	response.Body.Close()
	assertion.Nil(err)
	assertion.Equal("pong", string(body))
}

func Test_NewWithModeConfig(t *testing.T) {
	assertion := assert.New(t)

	app := fakeApp("development")
	assertion.Equal("for development", app.config.RunName())
}
