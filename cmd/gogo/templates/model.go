package templates

var (
	modelTemplate = `package models

import (
	"database/sql"
)

var (
	model *sql.Conn
)

// Config defines config for model driver
type Config struct {
	Host     string ` + "`" + `json:"host"` + "`" + `
	Username string ` + "`" + `json:"username"` + "`" + `
	Password string ` + "`" + `json:"password"` + "`" + `
}

// Setup inject model with driver conn
func Setup(config *Config) error {
	// TODO: create *sql.Conn with config

	model = &sql.Conn{}

	return nil
}
`
	modelTestTemplate = `package models

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
`
)
