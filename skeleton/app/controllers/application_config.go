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
	Model        *models.Config      `json:"model"`
	Domain       string              `json:"domain"`
	GettingStart *GettingStartConfig `json:"getting_start"`
	Debug        bool                `json:"debug"`
}

// NewAppConfig apply application config from *gogo.AppConfig
func NewAppConfig(config gogo.Configer) error {
	return config.UnmarshalJSON(&Config)
}

// Sample application config for illustration
type GettingStartConfig struct {
	Greeting string `json:"greeting"`
}
