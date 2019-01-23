package templates

var (
	applicationConfigTemplate = `package controllers

import (
	"github.com/dolab/gogo"

	"{{.Namespace}}/{{.Application}}/app/models"
)

var (
	Config *AppConfig
)

// AppConfig defines specs for config
type AppConfig struct {
	Model        *models.Config      ` + "`" + `json:"model"` + "`" + `
	Domain       string              ` + "`" + `json:"domain"` + "`" + `
	GettingStart *GettingStartConfig ` + "`" + `json:"getting_start"` + "`" + `
}

// NewAppConfig creates Config from gogo.Configer
func NewAppConfig(config gogo.Configer) error {
	return config.UnmarshalJSON(&Config)
}

// Sample config for illustration
type GettingStartConfig struct {
	Greeting string ` + "`" + `json:"greeting"` + "`" + `
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
	applicationConfigJSONTemplate = `{
	"mode": "test",
	"name": "{{.Application}}",
	"sections": {
		"development": {
			"server": {
				"addr": "localhost",
				"port": 9090,
				"healthz": true,
				"ssl": false,
				"request_timeout": 3,
				"response_timeout": 10,
				"request_id": "X-Request-Id"
			},
			"logger": {
				"output": "stderr",
				"level": "debug",
				"filter_params": ["password", "password_confirmation"]
			},
			"domain": "https://example.com",
			"getting_start": {
				"greeting": "Hello, gogo!"
			}
		},

		"test": {
			"server": {
				"addr": "localhost",
				"port": 9090,
				"ssl": false,
				"request_timeout": 3,
				"response_timeout": 10,
				"request_id": "X-Request-Id"
			},
			"logger": {
				"output": "nil",
				"level": "info",
				"filter_params": ["password", "password_confirmation"]
			},
			"domain": "https://example.com",
			"getting_start": {
				"greeting": "Hello, gogo!"
			}
		},

		"production": {
			"server": {
				"addr": "localhost",
				"port": 9090,
				"healthz": true,
				"ssl": true,
				"ssl_cert": "/path/to/ssl/cert",
				"ssl_key": "/path/to/ssl/key",
				"throttle": 3000,
				"demotion": 6000,
				"request_timeout": 3,
				"response_timeout": 10,
				"request_id": "X-Request-Id"
			},
			"logger": {
				"output": "stderr",
				"level": "warn",
				"filter_params": ["password", "password_confirmation"]
			}
		}
	}
}
`
)
