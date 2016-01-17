package controllers

import (
	"testing"
)

func Test_ExampleHello(t *testing.T) {
	testClient.Get(t, "/@getting_start/hello")

	testClient.AssertOK()
	testClient.AssertContains(Config.GettingStart.Greeting)
}
