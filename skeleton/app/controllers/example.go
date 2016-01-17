package controllers

import (
	"github.com/dolab/gogo"
)

var (
	Example *_Example
)

type _Example struct{}

// @route GET /@example/hello
func (_ *_Example) Hello(ctx *gogo.Context) {
	ctx.Logger.Warnf("Visiting domain is: %s", Config.Domain)

	ctx.Text(Config.Example.Greeting)
}
