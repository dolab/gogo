package controllers

import (
	"testing"

	"github.com/golib/assert"
)

func Test_AppConfig(t *testing.T) {
	it := assert.New(t)

	it.NotEmpty(Config.Domain)
	it.NotNil(Config.GettingStart)
}
