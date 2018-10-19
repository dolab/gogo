package controllers

import (
	"testing"
)

func Test_GettingStart_Hello(t *testing.T) {
	request := gogotest.New(t)
	request.Get("/@getting_start/hello")

	request.AssertOK()
	request.AssertContains(Config.GettingStart.Greeting)
}
