package gogo

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/golib/assert"
)

var (
	fakePort   uint64 = 9090
	fakeConfig        = func(name string) (*AppConfig, error) {
		root, _ := os.Getwd()

		data, err := ioutil.ReadFile(path.Join(root, "testdata", "config", name))
		if err != nil {
			panic(err)
		}

		port := atomic.AddUint64(&fakePort, 1)

		return NewAppConfigFromString(strings.Replace(string(data), "9090", strconv.FormatUint(port, 10), -1))
	}
)

func Test_NewConfig(t *testing.T) {
	it := assert.New(t)

	config, err := fakeConfig("application.yml")
	if it.Nil(err) {
		it.Implements((*Configer)(nil), config)

		it.Equal(Test, config.Mode)
		it.Equal("gogo", config.Name)
		it.NotEmpty(config.Sections)
	}
}

func Test_NewConfigWithoutMode(t *testing.T) {
	it := assert.New(t)

	config, err := NewAppConfigFromString(`{"name": "testing"}`)
	if it.Nil(err) {
		it.Equal(Development, config.Mode)
	}
}

func Test_NewConfigWithoutName(t *testing.T) {
	it := assert.New(t)

	config, err := NewAppConfigFromString(`{"mode": "test"}`)
	if it.Nil(err) {
		it.Equal("GOGO", config.Name)
	}
}

func Test_ConfigSetMode(t *testing.T) {
	it := assert.New(t)

	config, err := fakeConfig("application.yml")
	if it.Nil(err) {
		it.Equal(Test, config.Mode)

		config.SetMode(Production)
		it.Equal(Production, config.Mode)
	}
}

func Test_ConfigSection(t *testing.T) {
	it := assert.New(t)

	config, err := fakeConfig("application.yml")
	if it.Nil(err) {
		section := config.Section()
		it.NotNil(section.Server)
		it.NotNil(section.Logger)
	}
}

func Test_ConfigUnmarshalYAML(t *testing.T) {
	it := assert.New(t)

	var testConfig struct {
		GettingStart struct {
			Greeting string `yaml:"greeting"`
		} `yaml:"getting_start"`
	}

	config, err := fakeConfig("application.yml")
	if it.Nil(err) {
		config.SetMode(Development)

		err := config.UnmarshalYAML(&testConfig)
		if it.Nil(err) {
			it.Equal("Hello, gogo!", testConfig.GettingStart.Greeting)
		}
	}
}

func Test_ConfigWithDefaults(t *testing.T) {
	it := assert.New(t)

	config, err := NewAppConfigFromDefault()
	if it.Nil(err) {
		// it should work for development mode
		testModes := []RunMode{
			Development, Test, Production,
		}

		for _, mode := range testModes {
			config.SetMode(mode)

			section := config.Section()
			it.Equal(DefaultServerConfig.Addr, section.Server.Addr)
			it.Equal(DefaultServerConfig.Port, section.Server.Port)
			it.Equal(DefaultServerConfig.Ssl, section.Server.Ssl)
			it.Equal(DefaultLoggerConfig.Output, section.Logger.Output)
			it.Equal(DefaultLoggerConfig.LevelName, section.Logger.LevelName)
			it.Equal(DefaultLoggerConfig.FilterFields, section.Logger.FilterFields)
		}
	}
}
