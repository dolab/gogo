package models

import (
	"testing"

	"github.com/golib/assert"
)

func Test_Setup(t *testing.T) {
	assertion := assert.New(t)

	assertion.Nil(model)
	Setup(&Config{})
	assertion.NotNil(model)
}
