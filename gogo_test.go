package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	newMockApp = func() *AppServer {
		root, _ := os.Getwd()

		return New("test", root)
	}
)

func Test_New(t *testing.T) {
	assertion := assert.New(t)

	app := newMockApp()
	app.GET("/testing", func(ctx *Context) {
		ctx.Text("Hello, gogo!")
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	response, err := http.Get(ts.URL + "/testing")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(response.Body)
	assertion.Nil(err)
	assertion.Equal("Hello, gogo!", string(body))
}

func Test_NewWithModeConfig(t *testing.T) {
	assertion := assert.New(t)
	root, _ := os.Getwd()

	app := New("development", root)
	assertion.Equal("development testing", app.config.Name)
}
