package controllers

import (
	"github.com/dolab/gogo"
)

var (
	Config *AppConfig
)

// Application configuration specs
type AppConfig struct {
	Domain       string              `json:"domain"`
	GettingStart *GettingStartConfig `json:"getting_start"`
}

// NewAppConfig apply application config from *gogo.AppConfig
func NewAppConfig(config *gogo.AppConfig) error {
	return config.UnmarshalJSON(&Config)
}

// Sample application config for illustration
type GettingStartConfig struct {
	Greeting string `json:"greeting"`
}
