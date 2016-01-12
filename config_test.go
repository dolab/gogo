package gogo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	newMockConfig = func(name string) (*AppConfig, error) {
		root, _ := os.Getwd()

		return NewAppConfig(root + "/config/" + name)
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

	config, _ := newMockConfig("nomode.json")
	assertion.Equal(Development, config.Mode)
}

func Test_NewConfigWithoutName(t *testing.T) {
	assertion := assert.New(t)

	config, _ := newMockConfig("noname.json")
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
	assertion.NotNil(section)
}

func Test_ConfigUnmarshalJSON(t *testing.T) {
	assertion := assert.New(t)
	config, _ := newMockConfig("application.json")
	config.SetMode(Development)

	type appConfig struct {
		App struct {
			Key    string `json:"access_key"`
			Secret string `json:"access_secret"`
		} `json:"app"`
	}

	var testCase *appConfig

	err := config.UnmarshalJSON(&testCase)
	assertion.Nil(err)
	assertion.Equal("KEY", testCase.App.Key)
	assertion.Equal("SECRET", testCase.App.Secret)
}
