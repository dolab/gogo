package models

import (
	"database/sql"
	"testing"

	"github.com/golib/assert"
)

func Test_Setup(t *testing.T) {
	assertion := assert.New(t)

	assertion.Nil(model)
	Setup(&sql.Conn{})
	assertion.NotNil(model)
}
