package gogo

import (
	"crypto"
	"encoding/xml"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golib/assert"
)

type fakeContextStatusCoder struct {
	code int
}

func (statusCoder fakeContextStatusCoder) StatusCode() int {
	return statusCoder.code
}

func Test_NewContext(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.NotNil(ctx.Response)
	it.Nil(ctx.Request)
	it.Nil(ctx.Params)
	it.Nil(ctx.Logger)
	it.EqualValues(-1, ctx.cursor)
}

func Test_Context_Controller(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Empty(ctx.Controller())
}

func Test_Context_Action(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Empty(ctx.Action())
}

func Test_ContextWithSettings(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Equal(0, len(ctx.settings))
	it.Equal(0, len(ctx.frozenSettings))

	// set
	ctx.Set("middlewareKey", "middlewareValue")
	it.Equal(1, len(ctx.settings))
	it.Equal(0, len(ctx.frozenSettings))

	// get
	value, ok := ctx.Get("middlewareKey")
	if it.True(ok) {
		it.Equal("middlewareValue", value)
	}

	// get with empty
	value, ok = ctx.Get("unknownSetting")
	if it.False(ok) {
		it.Empty(value)
	}

	// MustGet
	it.Equal("middlewareValue", ctx.MustGet("middlewareKey"))

	// MustGet with empty
	it.Panics(func() {
		ctx.MustGet("unknownSetting")
	})
}

func Test_ContextWithFrozenSettings(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Equal(0, len(ctx.settings))
	it.Equal(0, len(ctx.frozenSettings))

	// final set
	err := ctx.SetFinal("middlewareFinalKey", "middlewareFinalValue")
	if it.Nil(err) {
		it.Equal(0, len(ctx.settings))
		it.Equal(1, len(ctx.frozenSettings))
	}

	// final get
	value, ok := ctx.GetFinal("middlewareFinalKey")
	if it.True(ok) {
		it.Equal("middlewareFinalValue", value)
	}

	// final set with conflict
	err = ctx.SetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	if it.NotNil(err) {
		it.EqualError(ErrSettingsKey, err.Error())
		it.Equal(1, len(ctx.frozenSettings))
	}

	value, ok = ctx.GetFinal("middlewareFinalKey")
	if it.True(ok) {
		it.Equal("middlewareFinalValue", value)
	}

	// final get with empty
	value, ok = ctx.GetFinal("unknownFinalSetting")
	if it.False(ok) {
		it.Empty(value)
	}

	// MustSetFinal
	ctx.MustSetFinal("mustMiddlewareFinalKey", "mustMiddlewareFinalValue")

	// MustSetFinal with conflict
	it.Panics(func() {
		ctx.MustSetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	})

	// MustGetFinal
	it.Equal("mustMiddlewareFinalValue", ctx.MustGetFinal("mustMiddlewareFinalKey"))

	// MustGetFinal with empty
	it.Panics(func() {
		ctx.MustGetFinal("unknownMiddlewareFinalKey")
	})
}

func Test_Context_RequestID(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Empty(ctx.RequestID())

	ctx.Logger = NewAppLogger("nil", "").New("requestID")
	it.Equal("requestID", ctx.RequestID())
}

func Test_Context_RequestURI(t *testing.T) {
	it := assert.New(t)

	ctx := NewContext()
	it.Panics(func() {
		ctx.RequestURI()
	})

	var err error

	ctx.Request, err = http.NewRequest(http.MethodGet, "https://example.com/path/to/resource", nil)
	if it.Nil(err) {
		it.Equal("/path/to/resource", ctx.RequestURI())
	}

	ctx.Request, err = http.NewRequest(http.MethodGet, "https://example.com/path/中文/resource", nil)
	if it.Nil(err) {
		it.Equal("/path/中文/resource", ctx.RequestURI())
	}
}

func Test_ContextWithHeader(t *testing.T) {
	it := assert.New(t)

	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	request.Header.Add("X-Canonical-Key", "Canonical-Value")
	request.Header["x-normal-key"] = []string{"normal value"}

	ctx := NewContext()
	ctx.Request = request

	// HasRawHeader
	it.True(ctx.HasRawHeader("X-Canonical-Key"))
	it.False(ctx.HasRawHeader("x-canonical-key"))
	it.True(ctx.HasRawHeader("x-normal-key"))
	it.False(ctx.HasRawHeader("X-Normal-Key"))

	// HasHeader
	it.True(ctx.HasHeader("X-Canonical-Key"))
	it.True(ctx.HasHeader("x-canonical-key"))
	it.False(ctx.HasHeader("x-normal-key"))
	it.False(ctx.HasHeader("X-Normal-Key"))

	// RawHeader
	it.Equal("Canonical-Value", ctx.RawHeader("X-Canonical-Key"))
	it.Empty(ctx.RawHeader("x-canonical-key"))
	it.Equal("normal value", ctx.RawHeader("x-normal-key"))
	it.Empty(ctx.RawHeader("X-Normal-Key"))

	// Header
	it.Equal("Canonical-Value", ctx.Header("X-Canonical-Key"))
	it.Equal("Canonical-Value", ctx.Header("x-canonical-key"))
	it.Empty(ctx.Header("x-normal-key"))
	it.Empty(ctx.Header("X-Normal-Key"))
}

func Test_Context_AddHeader(t *testing.T) {
	it := assert.New(t)

	recorder := httptest.NewRecorder()
	it.Empty(recorder.Header())

	ctx := NewContext()
	ctx.Response.Hijack(recorder)

	ctx.AddHeader("key", "value")
	it.NotEmpty(recorder.Header())
	it.Equal("value", recorder.Header().Get("key"))

	// more
	ctx.AddHeader("key", "value2")
	it.NotEmpty(recorder.Header())
	it.Equal("value", recorder.Header().Get("key"))
}

func Test_Context_SetHeader(t *testing.T) {
	it := assert.New(t)

	recorder := httptest.NewRecorder()
	it.Empty(recorder.Header())

	ctx := NewContext()
	ctx.Response.Hijack(recorder)

	ctx.SetHeader("key", "value")
	it.NotEmpty(recorder.Header())
	it.Equal("value", recorder.Header().Get("key"))

	// more
	ctx.SetHeader("key", "value2")
	it.NotEmpty(recorder.Header())
	it.Equal("value2", recorder.Header().Get("key"))
}

func Test_Context_SetStatus(t *testing.T) {
	it := assert.New(t)

	recorder := httptest.NewRecorder()
	it.Equal(http.StatusOK, recorder.Code)

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	it.Equal(http.StatusOK, ctx.Response.Status())

	ctx.SetStatus(http.StatusAccepted)
	it.Equal(http.StatusOK, recorder.Code)
	it.Equal(http.StatusAccepted, ctx.Response.Status())

	// more
	ctx.SetStatus(http.StatusOK)
	it.Equal(http.StatusOK, recorder.Code)
	it.Equal(http.StatusOK, ctx.Response.Status())
}

func Test_Context_Redirect(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.Hijack(recorder)

	ctx.Redirect(location)

	it.Equal(http.StatusFound, recorder.Code)
	it.Equal(location, recorder.Header().Get("Location"))
	it.EqualValues(math.MaxInt8, ctx.cursor)
}

func Test_Context_RedirectWithAbort(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request
	ctx.Logger = NewAppLogger("nil", "")
	ctx.middlewares = []Middleware{
		func(ctx *Context) {
			ctx.Redirect(location)

			ctx.Next()
		},
		func(ctx *Context) {
			ctx.Render(NewDefaultRender(ctx.Response), "next render")
		},
	}
	ctx.Next()

	it.Equal(location, recorder.Header().Get("Location"))
	it.NotContains(recorder.Body.String(), "next render")
}

func Test_Context_Return(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request

	// return with sample string
	s := "Hello, world!"

	err := ctx.Return(s)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(RenderDefaultContentType, recorder.Header().Get("Content-Type"))
		it.Equal(s, recorder.Body.String())
	}
}

func Test_Context_ReturnWithJson(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	request.Header.Set("Accept", "application/json, text/xml; charset=utf-8")

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request

	// return with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	err := ctx.Return(data)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("application/json", recorder.Header().Get("Content-Type"))
		it.Equal(`{"Name":"gogo","Age":5}`, strings.TrimSpace(recorder.Body.String()))
	}
}

func Test_Context_ReturnWithXml(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	request.Header.Set("Accept", "appication/json, text/xml; charset=utf-8")

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request

	// render with complex data type
	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	err := ctx.Return(data)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("text/xml", recorder.Header().Get("Content-Type"))
		it.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>", strings.TrimSpace(recorder.Body.String()))
	}
}

func Test_Context_Render(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}

	testCases := map[string]struct {
		w           Render
		contentType string
		data        interface{}
	}{
		"default render": {
			NewDefaultRender(ctx.Response),
			RenderDefaultContentType,
			"default render",
		},
		"hashed render": {
			NewHashRender(ctx.Response, crypto.MD5),
			RenderDefaultContentType,
			"hashed render",
		},
		"text render": {
			NewTextRender(ctx.Response),
			RenderDefaultContentType,
			"text render",
		},
		`{"success":false,"error":"not found"}`: {
			NewJsonRender(ctx.Response),
			"application/json",
			struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
			}{false, "not found"},
		},
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Result><Success>true</Success><Data>123</Data></Result>": {
			NewXmlRender(ctx.Response),
			"text/xml",
			struct {
				XMLName xml.Name `xml:"Result"`
				Success bool     `xml:"Success"`
				Data    int      `xml:"Data"`
			}{
				Success: true,
				Data:    123,
			},
		},
		"js_callback({\"success\":true,\"data\":123}\n);": {
			NewJsonpRender(ctx.Response, "js_callback"),
			"application/javascript",
			struct {
				Success bool `json:"success"`
				Data    int  `json:"data"`
			}{true, 123},
		},
	}
	for expected, testCase := range testCases {
		recorder.HeaderMap = http.Header{}
		recorder.Body.Reset()

		err := ctx.Render(testCase.w, testCase.data)
		if it.Nil(err) {
			it.Equal(testCase.contentType, recorder.Header().Get("Content-Type"))
			it.Equal(expected, strings.TrimSpace(recorder.Body.String()))
		}
	}
}

func Test_Context_RenderWithStatusCoder(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	statusCoder := fakeContextStatusCoder{
		code: 123,
	}

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request
	ctx.Logger = NewAppLogger("nil", "")

	err := ctx.Render(NewDefaultRender(ctx.Response), statusCoder)
	if it.Nil(err) {
		it.Equal(123, ctx.Response.Status())
	}
}

func Test_Context_RenderWithAbort(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Response.Hijack(recorder)
	ctx.Request = request
	ctx.Logger = NewAppLogger("nil", "")
	ctx.middlewares = []Middleware{
		func(ctx *Context) {
			ctx.Render(NewDefaultRender(ctx.Response), "render")

			ctx.Next()
		},
		func(ctx *Context) {
			ctx.Render(NewDefaultRender(ctx.Response), "next render")
		},
	}
	ctx.Next()

	it.Equal("render", recorder.Body.String())
	it.EqualValues(-128, ctx.cursor)
}

func Test_Context_Next(t *testing.T) {
	it := assert.New(t)

	counter := 0
	middleware1 := func(ctx *Context) {
		counter++

		ctx.Next()
	}
	middleware2 := func(ctx *Context) {
		counter++

		ctx.Next()
	}

	ctx := NewContext()
	ctx.Logger = NewAppLogger("nil", "")
	ctx.middlewares = append(ctx.middlewares, middleware1, middleware2)
	ctx.Next()

	it.EqualValues(2, ctx.cursor)
	it.Equal(2, counter)
}

func Test_Context_Abort(t *testing.T) {
	it := assert.New(t)

	counter := 0
	middleware1 := func(ctx *Context) {
		counter++

		ctx.Next()
	}
	middleware2 := func(ctx *Context) {
		counter += 2

		ctx.Next()
	}

	ctx := NewContext()
	ctx.Logger = NewAppLogger("nil", "")
	ctx.middlewares = append(ctx.middlewares, middleware1, middleware2)
	ctx.Abort()
	ctx.Next()

	it.EqualValues(-128, ctx.cursor)
	it.Equal(0, counter)
}

func Benchmark_Context(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	ctx := NewContext()
	ctx.Request = httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	ctx.Logger = NewAppLogger("nil", "")

	for i := 0; i < b.N; i++ {
		ctx.run(recorder, nil, nil)
	}
}
