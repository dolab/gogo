package gogo

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	newMockConfig = func(name string) (*AppConfig, error) {
		root, _ := os.Getwd()

		return NewAppConfig(path.Join(root, "skeleton", "config", name))
	}
)

func Test_NewConfig(t *testing.T) {
	assertion := assert.New(t)

	config, err := newMockConfig("application.json")
	assertion.Nil(err)
	assertion.Equal(Test, config.Mode)
	assertion.Equal("gogo", config.Name)
	assertion.NotEmpty(config.Sections)
}

func Test_NewConfigWithoutMode(t *testing.T) {
	assertion := assert.New(t)
	config, _ := NewStringAppConfig(`{
    "name": "testing"
}`)

	assertion.Equal(Development, config.Mode)
}

func Test_NewConfigWithoutName(t *testing.T) {
	assertion := assert.New(t)
	config, _ := NewStringAppConfig(`{
    "mode": "test"
}`)

	assertion.Equal("GOGO", config.Name)
}

func Test_ConfigSetMode(t *testing.T) {
	assertion := assert.New(t)

	config, _ := newMockConfig("application.json")
	assertion.Equal(Test, config.Mode)

	config.SetMode(Production)
	assertion.Equal(Production, config.Mode)
}

func Test_ConfigSection(t *testing.T) {
	assertion := assert.New(t)
	config, _ := newMockConfig("application.json")

	section := config.Section()
	assertion.NotNil(section.Server)
	assertion.NotNil(section.Logger)
}

func Test_ConfigUnmarshalJSON(t *testing.T) {
	var testConfig struct {
		GettingStart struct {
			Greeting string `json:"greeting"`
		} `json:"getting_start"`
	}

	assertion := assert.New(t)
	config, _ := newMockConfig("application.json")
	config.SetMode(Development)

	err := config.UnmarshalJSON(&testConfig)
	assertion.Nil(err)
	assertion.Equal("Hello, gogo!", testConfig.GettingStart.Greeting)
}
