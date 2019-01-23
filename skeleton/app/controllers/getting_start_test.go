package controllers

import (
	"testing"
)

func Test_GettingStart_Hello(t *testing.T) {
	request := gogotesting.New(t)
	request.Get("/@gretting/hello")

	request.AssertOK()
	request.AssertContains(Config.GettingStart.Greeting)
}
