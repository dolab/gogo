package gogo

const (
	Development RunMode = "development"
	Test        RunMode = "test"
	Production  RunMode = "production"
)

const (
	DefaultMaxMultiformBytes = 32 << 20 // 32M
	DefaultMaxHeaderBytes    = 64 << 10 // 64k
)

const (
	DefaultHttpRequestID       = "X-Request-Id"
	DefaultMaxHttpRequestIDLen = 32
	DefaultHttpRequestTimeout  = 30 // 30s
	DefaultHttpResponseTimeout = 30 // 30s
)

const (
	RenderDefaultContentType = "text/plain; charset=utf-8"
	RenderJsonContentType    = "application/json"
	RenderJsonPContentType   = "application/javascript"
	RednerXmlContentType     = "text/xml"
)

// RunMode defines app run mode
type RunMode string

// IsValid returns true if mode is valid
func (mode RunMode) IsValid() bool {
	switch mode {
	case Development, Test, Production:
		return true
	}

	return false
}

// IsDevelopment returns true if mode is development
func (mode RunMode) IsDevelopment() bool {
	return mode == Development
}

// IsTest returns true if mode is test
func (mode RunMode) IsTest() bool {
	return mode == Test
}

// IsProduction returns true if mode equals to production
func (mode RunMode) IsProduction() bool {
	return mode == Production
}

func (mode RunMode) String() string {
	return string(mode)
}
