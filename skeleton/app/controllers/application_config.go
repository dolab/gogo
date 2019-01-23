package controllers

import (
	"github.com/dolab/gogo"

	"github.com/skeleton/app/models"
)

var (
	Config *AppConfig
)

// Application configuration specs
type AppConfig struct {
	Domain       string              `json:"domain"`
	Model        *models.Config      `json:"model"`
	GettingStart *GettingStartConfig `json:"getting_start"`
}

// NewAppConfig apply application config from *gogo.AppConfig
func NewAppConfig(config gogo.Configer) error {
	return config.UnmarshalJSON(&Config)
}

// Sample application config for illustration
type GettingStartConfig struct {
	Greeting string `json:"greeting"`
}
