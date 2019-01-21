package gogo

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/dolab/logger"
)

var (
	DefaultServerConfig = &ServerConfig{
		Addr:      "127.0.0.1",
		Port:      9090,
		RTimeout:  10, // 10s
		WTimeout:  10, // 10s
		Ssl:       false,
		RequestID: DefaultRequestIDKey,
	}

	DefaultLoggerConfig = &LoggerConfig{
		Output:       "stderr",
		LevelName:    "info",
		FilterFields: []string{"password", "token"},
	}

	DefaultSectionConfig = &SectionConfig{
		Server: DefaultServerConfig,
		Logger: DefaultLoggerConfig,
	}
)

// AppConfig defines config component of gogo.
// It implements Configer interface.
type AppConfig struct {
	Mode RunMode `json:"mode"`
	Name string  `json:"name"`

	Sections map[RunMode]*json.RawMessage `json:"sections"`
}

// NewAppConfig returns *AppConfig by parsing application.json
func NewAppConfig(filename string) (*AppConfig, error) {
	if strings.HasPrefix(filename, GogoSchema) {
		return NewAppConfigFromDefault()
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewAppConfigFromString(string(b))
}

// NewAppConfigFromString returns *AppConfig by parsing json string
func NewAppConfigFromString(s string) (*AppConfig, error) {
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

// NewAppConfigFromDefault returns *AppConfig of defaults
func NewAppConfigFromDefault() (*AppConfig, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"mode": Development,
		"name": "gogo",
		"sections": map[RunMode]interface{}{
			Development: DefaultSectionConfig,
			Test:        DefaultSectionConfig,
			Production:  DefaultSectionConfig,
		},
	})

	return NewAppConfigFromString(string(data))
}

// RunMode returns the current mode of *AppConfig
func (config *AppConfig) RunMode() RunMode {
	return config.Mode
}

// RunName returns the application name
func (config *AppConfig) RunName() string {
	return config.Name
}

// SetMode changes config mode
func (config *AppConfig) SetMode(mode RunMode) {
	if !mode.IsValid() {
		return
	}

	config.Mode = mode
}

// Section is shortcut of retreving app server and logger configurations at once.
// It returns SectionConfig if exists, otherwise returns DefaultSectionConfig instead.
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

// SectionConfig defines config spec for internal usage
type SectionConfig struct {
	Server *ServerConfig `json:"server"`
	Logger *LoggerConfig `json:"logger"`
}

// ServerConfig defines config spec of AppServer
type ServerConfig struct {
	Addr           string `json:"addr"`             // listen address
	Port           int    `json:"port"`             // listen port
	RTimeout       int    `json:"request_timeout"`  // unit in second
	WTimeout       int    `json:"response_timeout"` // unit in second
	MaxHeaderBytes int    `json:"max_header_bytes"` // unit in byte

	HTTP2   bool   `json:"http2"` // use http2 server
	Ssl     bool   `json:"ssl"`
	SslCert string `json:"ssl_cert"`
	SslKey  string `json:"ssl_key"`

	Throttle  int    `json:"throttle"` // in time.Second/throttle ms
	Slowdown  int    `json:"slowdown"`
	RequestID string `json:"request_id"`
}

// LoggerConfig defines config spec of AppLogger
type LoggerConfig struct {
	Output    string `json:"output"` // valid values [stdout|stderr|null|path/to/file]
	LevelName string `json:"level"`  // valid values [debug|info|warn|error]

	FilterFields []string `json:"filter_fields"` // sensitive fields which should filter out when logging
}

// Level returns logger.Level by its name
func (l *LoggerConfig) Level() logger.Level {
	return logger.ResolveLevelByName(l.LevelName)
}
