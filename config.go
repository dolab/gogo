package gogo

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dolab/logger"
)

var (
	DefaultServerConfig = &ServerConfig{
		Addr:      "127.0.0.1",
		Port:      9090,
		Ssl:       false,
		RequestId: DefaultHttpRequestId,
	}

	DefaultLoggerConfig = &LoggerConfig{
		Output:    "stderr",
		LevelName: "info",
	}

	DefaultSectionConfig = &SectionConfig{
		Server: DefaultServerConfig,
		Logger: DefaultLoggerConfig,
	}
)

type AppConfig struct {
	Mode RunMode `json:"mode"`
	Name string  `json:"name"`

	Sections map[RunMode]*json.RawMessage `json:"sections"`
}

// NewAppConfig returns app config by parsing application.json
func NewAppConfig(filename string) (*AppConfig, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewStringAppConfig(string(b))
}

// NewStringAppConfig returns app config by parsing json string
func NewStringAppConfig(s string) (*AppConfig, error) {
	var config *AppConfig

	err := json.Unmarshal([]byte(s), &config)
	if err != nil {
		return nil, err
	}

	// default to development mode
	if !config.Mode.IsValid() {
		config.SetMode(Development)
	}

	// adjust app name
	if config.Name == "" {
		config.Name = "GOGO"
	}

	return config, nil
}

// SetMode changes config mode
func (config *AppConfig) SetMode(mode RunMode) {
	if !mode.IsValid() {
		return
	}

	config.Mode = mode
}

// Section is shortcut of retreving app server and logger configurations at one time
// It returns SectionConfig if exists, otherwise returns DefaultSectionConfig instead
func (config *AppConfig) Section() *SectionConfig {
	var sconfig *SectionConfig

	err := config.UnmarshalJSON(&sconfig)
	if err != nil {
		return DefaultSectionConfig
	}

	return sconfig
}

// UnmarshalJSON parses JSON-encoded data of section and stores the result in the value pointed to by v.
// It returns ErrConfigSection error if section of the current mode does not exist.
func (config *AppConfig) UnmarshalJSON(v interface{}) error {
	section, ok := config.Sections[config.Mode]
	if !ok {
		return ErrConfigSection
	}

	return json.Unmarshal([]byte(*section), &v)
}

// app server and logger config
type SectionConfig struct {
	Server *ServerConfig `json:"server"`
	Logger *LoggerConfig `json:"logger"`
}

// server config spec
type ServerConfig struct {
	Addr           string `json:"addr"`
	Port           int    `json:"port"`
	RTimeout       int    `json:"request_timeout"`
	WTimeout       int    `json:"response_timeout"`
	MaxHeaderBytes int    `json:"max_header_bytes"`

	Ssl     bool   `json:"ssl"`
	SslCert string `json:"ssl_cert"`
	SslKey  string `json:"ssl_key"`

	RequestId string `json:"request_id"`
}

// logger config spec
type LoggerConfig struct {
	Output    string `json:"output"` // valid values [stdout|stderr|null|path/to/file]
	LevelName string `json:"level"`  // valid values [debug|info|warn|error]

	FilterParams []string `json:"filter_params"`
}

func (l *LoggerConfig) Level() logger.Level {
	return logger.ResolveLevelByName(l.LevelName)
}
