package gogo

import (
	"os"
	"path"
	"testing"

	"github.com/golib/assert"
)

var (
	fakeConfig = func(name string) (*AppConfig, error) {
		root, _ := os.Getwd()

		return NewAppConfig(path.Join(root, "skeleton", "gogo", "config", name))
	}
)

func Test_NewConfig(t *testing.T) {
	assertion := assert.New(t)

	config, err := fakeConfig("application.json")
	assertion.Nil(err)
	assertion.Equal(Test, config.Mode)
	assertion.Equal("gogo", config.Name)
	assertion.NotEmpty(config.Sections)
	assertion.Implements((*Configer)(nil), config)
}

func Test_NewConfigWithoutMode(t *testing.T) {
	assertion := assert.New(t)
	config, _ := NewAppConfigFromString(`{
    "name": "testing"
}`)

	assertion.Equal(Development, config.Mode)
}

func Test_NewConfigWithoutName(t *testing.T) {
	assertion := assert.New(t)
	config, _ := NewAppConfigFromString(`{
    "mode": "test"
}`)

	assertion.Equal("GOGO", config.Name)
}

func Test_ConfigSetMode(t *testing.T) {
	assertion := assert.New(t)

	config, _ := fakeConfig("application.json")
	assertion.Equal(Test, config.Mode)

	config.SetMode(Production)
	assertion.Equal(Production, config.Mode)
}

func Test_ConfigSection(t *testing.T) {
	assertion := assert.New(t)
	config, _ := fakeConfig("application.json")

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
	config, _ := fakeConfig("application.json")
	config.SetMode(Development)

	err := config.UnmarshalJSON(&testConfig)
	assertion.Nil(err)
	assertion.Equal("Hello, gogo!", testConfig.GettingStart.Greeting)
}
