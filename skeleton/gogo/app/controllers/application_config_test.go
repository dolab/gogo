package controllers

import (
	"testing"

	"github.com/golib/assert"
)

func Test_AppConfig(t *testing.T) {
	assertion := assert.New(t)

	assertion.NotEmpty(Config.Domain)
	assertion.NotNil(Config.GettingStart)
}
