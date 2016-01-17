package controllers

import (
	"testing"
)

func Test_ExampleHello(t *testing.T) {
	testClient.Get(t, "/@example/hello")

	testClient.AssertOK()
	testClient.AssertContains(Config.Example.Greeting)
}
