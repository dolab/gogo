package debugger

// A Config defines user custom settings of debugger
type Config struct {
	Priority          int  `yaml:"priority"`
	DebugRequest      bool `yaml:"debug_request"`
	DebugRequestBody  bool `yaml:"debug_request_body"`
	DebugResponse     bool `yaml:"debug_response"`
	DebugResponseBody bool `yaml:"debug_response_body"`
}

// Debugable returns true if enabled
func (c *Config) Debugable() bool {
	if c == nil {
		return false
	}

	return c.DebugRequest || c.DebugResponse
}
