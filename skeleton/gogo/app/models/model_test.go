package models

import (
	"testing"

	"github.com/golib/assert"
)

func Test_Setup(t *testing.T) {
	it := assert.New(t)

	it.Nil(model)
	Setup(&Config{})
	it.NotNil(model)
}
