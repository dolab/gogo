package templates

var (
	applicationConfigTemplate = `package controllers

import (
	"github.com/dolab/gogo"

	"{{.Namespace}}/{{.Application}}/app/models"
)

// Shared Config
var (
	Config *AppConfig
)

// AppConfig defines specs for config
type AppConfig struct {
	Model        *models.Config      ` + "`" + `yaml:"model"` + "`" + `
	Domain       string              ` + "`" + `yaml:"domain"` + "`" + `
	GettingStart *GettingStartConfig ` + "`" + `yaml:"getting_start"` + "`" + `
	Debug        bool                ` + "`" + `yaml:"debug"` + "`" + `
}

// NewAppConfig creates Config from gogo.Configer
func NewAppConfig(config gogo.Configer) error {
	return config.UnmarshalYAML(&Config)
}

// GettingStartConfig is a sample config for illustration
type GettingStartConfig struct {
	Greeting string ` + "`" + `yaml:"greeting"` + "`" + `
}
`
	applicationConfigTestTemplate = `package controllers

import (
	"testing"

	"github.com/golib/assert"
)

func Test_AppConfig(t *testing.T) {
	it := assert.New(t)

	it.NotEmpty(Config.Domain)
	it.NotNil(Config.GettingStart)
}
`
	applicationConfigYAMLTemplate = `---
mode: test
name: {{.Application}}

default_server: &default_server
	addr: localhost
	port: 9090
	ssl: false
	request_timeout: 3
	response_timeout: 10
	request_id: X-Request-Id

default_logger: &default_logger
	output: nil
	level: debug
	filter_params:
	- password
	- password_confirmation

sections:
	development:
		server:
			<<: *default_server
		logger:
			<<: *default_logger
		domain: https://example.com
		getting_start:
			greeting: Hello, gogo!
		debug: true
	test:
		server:
			<<: *default_server
			request_id: ''
		logger:
			<<: *default_logger
		domain: https://example.com
		getting_start:
			greeting: Hello, gogo!
		debug: false
	production:
		server:
			<<: *default_server
			ssl: true
			ssl_cert: "/path/to/ssl/cert"
			ssl_key: "/path/to/ssl/key"
		logger:
			<<: *default_logger	
`
)
