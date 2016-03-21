package gogo

const (
	Development RunMode = "development"
	Test        RunMode = "test"
	Production  RunMode = "production"

	DefaultMaxMultiformBytes = 32 << 20 // 32M
	DefaultMaxHeaderBytes    = 64 << 10 // 64k

	DefaultHttpRequestId       = "X-Request-Id"
	DefaultHttpRequestTimeout  = 30 // 30s
	DefaultHttpResponseTimeout = 30 // 30s
)

type RunMode string

func (mode RunMode) IsValid() bool {
	switch mode {
	case Development, Test, Production:
		return true
	}

	return false
}

func (mode RunMode) IsDevelopment() bool {
	return mode == Development
}

func (mode RunMode) IsTest() bool {
	return mode == Test
}

func (mode RunMode) IsProduction() bool {
	return mode == Production
}
