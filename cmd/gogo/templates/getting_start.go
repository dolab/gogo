package templates

var (
	gettingStartTemplate = `package controllers

import (
	"github.com/dolab/gogo"
)

// @resources GettingStart
var (
	GettingStart *_GettingStart
)

type _GettingStart struct{}

// @route GET /@greeting/hello
func (_ *_GettingStart) Hello(ctx *gogo.Context) {
	name := ctx.Params.Get("name")
	if name == "" {
		name = Config.GettingStart.Greeting
	}

	ctx.Text(name)
}
`
	gettingStartTestTemplate = `package controllers

import (
	"net/http"
	"net/url"
	"testing"
)

func Test_GettingStart_Hello(t *testing.T) {
	// it should work without greeting
	request := gogotesting.New(t)
	request.Get("/v1/@greeting/hello")

	request.AssertOK()
	request.AssertContains(Config.GettingStart.Greeting)

	// it should work with custom greeting
	greeting := "Hi, gogo!"

	params := url.Values{}
	params.Add("name", greeting)

	request = gogotesting.New(t)
	request.Get("/v1/@greeting/hello", params)

	request.AssertOK()
	request.AssertContains(greeting)

	// it should return 404 when not found
	request = gogotesting.New(t)
	request.Get("/@greeting/hello")

	request.AssertStatus(http.StatusNotFound)
}
`
)
