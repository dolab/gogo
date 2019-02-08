package gogo

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dolab/logger"
	yaml "gopkg.in/yaml.v2"
)

// default configurations
var (
	DefaultServerConfig = &ServerConfig{
		Addr:      "127.0.0.1",
		Port:      9090,
		RTimeout:  5,  // 5s
		WTimeout:  10, // 10s
		Ssl:       false,
		Healthz:   true,
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

	DefaultMiddlewareConfig = &MiddlewareConfig{}
)

// AppConfig defines config component of gogo.
// It implements Configer interface.
type AppConfig struct {
	Mode     RunMode                 `yaml:"mode"`
	Name     string                  `yaml:"name"`
	Sections map[RunMode]interface{} `yaml:"sections"`

	middlewares *MiddlewareConfig
	filepath    string
}

// NewAppConfig returns *AppConfig by parsing application.yml
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

// NewAppConfigFromString returns *AppConfig by parsing.yml string
func NewAppConfigFromString(s string) (*AppConfig, error) {
	var config *AppConfig

	err := yaml.Unmarshal([]byte(s), &config)
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
	data, _ := yaml.Marshal(map[string]interface{}{
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

// Section is shortcut of retreving app config by section at once.
// It returns SectionConfig if exists, otherwise returns DefaultSectionConfig instead.
func (config *AppConfig) Section() *SectionConfig {
	var section *SectionConfig

	err := config.UnmarshalYAML(&section)
	if err != nil {
		return DefaultSectionConfig
	}

	return section
}

// UnmarshalJSON parses JSON-encoded data of section and stores the result in the value pointed to by v.
// It returns ErrConfigSection error if section of the current mode does not exist.
func (config *AppConfig) UnmarshalJSON(v interface{}) error {
	return fmt.Errorf("DEPRECATED!!! Please use yaml format")
}

// UnmarshalYAML parses YAML-encoded data of section and stores the result in the value pointed to by v.
// It returns ErrConfigSection error if section of the current mode does not exist.
func (config *AppConfig) UnmarshalYAML(v interface{}) error {
	section, ok := config.Sections[config.Mode]
	if !ok {
		return ErrConfigSection
	}

	data, err := yaml.Marshal(section)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, v)
}

// Middlewares returns a MiddlewareConfiger wrapped with parsed YAML-encoded data of all middlewares.
func (config *AppConfig) Middlewares() MiddlewareConfiger {
	return config.middlewares
}

// LoadMiddlewares reads all config of middlewares
func (config *AppConfig) LoadMiddlewares() error {
	filename := FindMiddlewareConfigFile(config.filepath)

	if strings.HasPrefix(filename, GogoSchema) {
		config.middlewares = DefaultMiddlewareConfig
		return nil
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &config.middlewares)
}

// SectionConfig defines config spec for internal usage
type SectionConfig struct {
	Server *ServerConfig `yaml:"server"`
	Logger *LoggerConfig `yaml:"logger"`
}

// ServerConfig defines config spec of AppServer
type ServerConfig struct {
	Addr           string `yaml:"addr"`             // listen address
	Port           int    `yaml:"port"`             // listen port
	RTimeout       int    `yaml:"request_timeout"`  // unit in second
	WTimeout       int    `yaml:"response_timeout"` // unit in second
	MaxHeaderBytes int    `yaml:"max_header_bytes"` // unit in byte
	RequestID      string `yaml:"request_id"`

	// ssl support
	Ssl     bool   `yaml:"ssl"`
	SslCert string `yaml:"ssl_cert"`
	SslKey  string `yaml:"ssl_key"`

	HTTP2    bool `yaml:"http2"`    // enable http2
	Healthz  bool `yaml:"healthz"`  // enable /-/healthz
	Throttle int  `yaml:"throttle"` // in time.Second/throttle ms
	Demotion int  `yaml:"demotion"`
}

// MiddlewareConfig defines config spec of middleware
type MiddlewareConfig map[string]interface{}

// Unmarshal parses YAML-encoded data of defined with name and stores the result in the
// value pointed to by v. It returns error if there is no config data for the name.
func (config MiddlewareConfig) Unmarshal(name string, v interface{}) error {
	if config == nil {
		return nil
	}

	idata, ok := config[name]
	if !ok {
		return fmt.Errorf("no config data for middleware %q", name)
	}

	b, err := yaml.Marshal(idata)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, v)
}

// LoggerConfig defines config spec of AppLogger
type LoggerConfig struct {
	Output       string   `yaml:"output"`        // valid values [stdout|stderr|null|nil|path/to/file]
	LevelName    string   `yaml:"level"`         // valid values [debug|info|warn|error]
	FilterFields []string `yaml:"filter_fields"` // sensitive fields which should be filtered out when logging
}

// Level returns logger.Level by its name
func (l *LoggerConfig) Level() logger.Level {
	return logger.ResolveLevelByName(l.LevelName)
}
