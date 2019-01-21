package gogo

// run mode
const (
	Development RunMode = "development"
	Test        RunMode = "test"
	Production  RunMode = "production"
)

// gogo schema and internal route
const (
	GogoSchema  = "gogo://"
	GogoHealthz = "/-/healthz"
)

// server defaults
const (
	DefaultRequestIDKey    = "X-Request-Id"
	DefaultRequestIDMaxLen = 32
	DefaultRequestTimeout  = 10 // 10s
	DefaultResponseTimeout = 10 // 10s
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

type contextKey int

const (
	ctxLoggerKey contextKey = iota + 1
)
