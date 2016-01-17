package controllers

import (
	"github.com/dolab/gogo"
)

var (
	Config *AppConfig
)

// Application configuration specs
type AppConfig struct {
	Domain  string         `json:"domain"`
	Example *ExampleConfig `json:"example"`
}

// NewAppConfig apply application config from *gogo.AppConfig
func NewAppConfig(config *gogo.AppConfig) error {
	return config.UnmarshalJSON(&Config)
}

// Application database config
type ExampleConfig struct {
	Greeting string `json:"greeting"`
}
