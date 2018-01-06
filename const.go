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
