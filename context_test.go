package gogo

import (
	"crypto"
	"encoding/xml"
	"math"
	"net/http"
	"net/http/httptest"
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
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.NotNil(ctx.Response)
	assertion.Nil(ctx.Request)
	assertion.Nil(ctx.Params)
	assertion.Nil(ctx.Logger)
	assertion.EqualValues(-1, ctx.cursor)
}

func Test_Context_Controller(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Empty(ctx.Controller())
}

func Test_Context_Action(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Empty(ctx.Action())
}

func Test_ContextWithSettings(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Equal(0, len(ctx.settings))
	assertion.Equal(0, len(ctx.frozenSettings))

	// set
	ctx.Set("middlewareKey", "middlewareValue")
	assertion.Equal(1, len(ctx.settings))
	assertion.Equal(0, len(ctx.frozenSettings))

	// get
	value, ok := ctx.Get("middlewareKey")
	assertion.True(ok)
	assertion.Equal("middlewareValue", value)

	// get with empty
	value, ok = ctx.Get("unknownSetting")
	assertion.False(ok)
	assertion.Empty(value)

	// MustGet
	assertion.Equal("middlewareValue", ctx.MustGet("middlewareKey"))

	// MustGet with empty
	assertion.Panics(func() {
		ctx.MustGet("unknownSetting")
	})
}

func Test_ContextWithFrozenSettings(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Equal(0, len(ctx.settings))
	assertion.Equal(0, len(ctx.frozenSettings))

	// final set
	err := ctx.SetFinal("middlewareFinalKey", "middlewareFinalValue")
	assertion.Nil(err)
	assertion.Equal(0, len(ctx.settings))
	assertion.Equal(1, len(ctx.frozenSettings))

	// final get
	value, ok := ctx.GetFinal("middlewareFinalKey")
	assertion.True(ok)
	assertion.Equal("middlewareFinalValue", value)

	// final set with conflict
	err = ctx.SetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	assertion.EqualError(err, ErrSettingsKey.Error())
	assertion.Equal(1, len(ctx.frozenSettings))

	value, ok = ctx.GetFinal("middlewareFinalKey")
	assertion.True(ok)
	assertion.Equal("middlewareFinalValue", value)

	// final get with empty
	value, ok = ctx.GetFinal("unknownFinalSetting")
	assertion.False(ok)
	assertion.Empty(value)

	// MustSetFinal
	ctx.MustSetFinal("mustMiddlewareFinalKey", "mustMiddlewareFinalValue")

	// MustSetFinal with conflict
	assertion.Panics(func() {
		ctx.MustSetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	})

	// MustGetFinal
	value = ctx.MustGetFinal("mustMiddlewareFinalKey")
	assertion.Equal("mustMiddlewareFinalValue", value)

	// MustGetFinal with empty
	assertion.Panics(func() {
		ctx.MustGetFinal("unknownMiddlewareFinalKey")
	})
}

func Test_Context_RequestID(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Empty(ctx.RequestID())

	ctx.Logger = NewAppLogger("stderr", "").New("requestID")
	assertion.Equal("requestID", ctx.RequestID())
}

func Test_Context_RequestURI(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.Panics(func() {
		ctx.RequestURI()
	})

	ctx.Request, _ = http.NewRequest(http.MethodGet, "https://example.com/path/to/resource", nil)
	assertion.Equal("/path/to/resource", ctx.RequestURI())

	ctx.Request, _ = http.NewRequest(http.MethodGet, "https://example.com/path/中文/resource", nil)
	assertion.Equal("/path/中文/resource", ctx.RequestURI())
}

func Test_ContextWithHeader(t *testing.T) {
	assertion := assert.New(t)

	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	request.Header.Add("X-Canonical-Key", "Canonical-Value")
	request.Header["x-normal-key"] = []string{"normal value"}

	ctx := NewContext()
	ctx.Request = request

	// HasRawHeader
	assertion.True(ctx.HasRawHeader("X-Canonical-Key"))
	assertion.False(ctx.HasRawHeader("x-canonical-key"))
	assertion.True(ctx.HasRawHeader("x-normal-key"))
	assertion.False(ctx.HasRawHeader("X-Normal-Key"))

	// HasHeader
	assertion.True(ctx.HasHeader("X-Canonical-Key"))
	assertion.True(ctx.HasHeader("x-canonical-key"))
	assertion.False(ctx.HasHeader("x-normal-key"))
	assertion.False(ctx.HasHeader("X-Normal-Key"))

	// RawHeader
	assertion.Equal("Canonical-Value", ctx.RawHeader("X-Canonical-Key"))
	assertion.Empty(ctx.RawHeader("x-canonical-key"))
	assertion.Equal("normal value", ctx.RawHeader("x-normal-key"))
	assertion.Empty(ctx.RawHeader("X-Normal-Key"))

	// Header
	assertion.Equal("Canonical-Value", ctx.Header("X-Canonical-Key"))
	assertion.Equal("Canonical-Value", ctx.Header("x-canonical-key"))
	assertion.Empty(ctx.Header("x-normal-key"))
	assertion.Empty(ctx.Header("X-Normal-Key"))
}

func Test_Context_AddHeader(t *testing.T) {
	assertion := assert.New(t)

	recorder := httptest.NewRecorder()
	assertion.Empty(recorder.Header())

	ctx := NewContext()
	ctx.Response.(*Response).reset(recorder)

	ctx.AddHeader("key", "value")
	assertion.NotEmpty(recorder.Header())
	assertion.Equal("value", recorder.Header().Get("key"))

	// more
	ctx.AddHeader("key", "value2")
	assertion.NotEmpty(recorder.Header())
	assertion.Equal("value", recorder.Header().Get("key"))
}

func Test_Context_SetHeader(t *testing.T) {
	assertion := assert.New(t)

	recorder := httptest.NewRecorder()
	assertion.Empty(recorder.Header())

	ctx := NewContext()
	ctx.Response.(*Response).reset(recorder)

	ctx.SetHeader("key", "value")
	assertion.NotEmpty(recorder.Header())
	assertion.Equal("value", recorder.Header().Get("key"))

	// more
	ctx.SetHeader("key", "value2")
	assertion.NotEmpty(recorder.Header())
	assertion.Equal("value2", recorder.Header().Get("key"))
}

func Test_Context_SetStatus(t *testing.T) {
	assertion := assert.New(t)

	recorder := httptest.NewRecorder()
	assertion.Equal(http.StatusOK, recorder.Code)

	ctx := NewContext()
	ctx.Response.(*Response).reset(recorder)
	assertion.Equal(http.StatusOK, ctx.Response.Status())

	ctx.SetStatus(http.StatusAccepted)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(http.StatusAccepted, ctx.Response.Status())

	// more
	ctx.SetStatus(http.StatusOK)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(http.StatusOK, ctx.Response.Status())
}

func Test_Context_Redirect(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)

	ctx.Redirect(location)

	assertion.Equal(http.StatusFound, recorder.Code)
	assertion.Equal(location, recorder.Header().Get("Location"))
	assertion.EqualValues(math.MaxInt8, ctx.cursor)
}

func Test_Context_RedirectWithAbort(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)
	ctx.Logger = NewAppLogger("stderr", "")
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

	assertion.Equal(location, recorder.Header().Get("Location"))
	assertion.NotContains(recorder.Body.String(), "next render")
}

func Test_Context_Return(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)

	// return with sample string
	s := "Hello, world!"

	err := ctx.Return(s)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(RenderDefaultContentType, recorder.Header().Get("Content-Type"))
	assertion.Equal(s, recorder.Body.String())
}

func Test_Context_ReturnWithJson(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	request.Header.Set("Accept", "application/json, text/xml; charset=utf-8")

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)

	// return with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	err := ctx.Return(data)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("application/json", recorder.Header().Get("Content-Type"))
	assertion.Contains(recorder.Body.String(), `{"Name":"gogo","Age":5}`)
}

func Test_Context_ReturnWithXml(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	request.Header.Set("Accept", "appication/json, text/xml; charset=utf-8")

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)

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
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("text/xml", recorder.Header().Get("Content-Type"))
	assertion.Contains(recorder.Body.String(), `<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>`)
}

func Test_Context_Render(t *testing.T) {
	assertion := assert.New(t)
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
		`js_callback({"success":true,"data":123});`: {
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
		assertion.Nil(err)
		assertion.Equal(testCase.contentType, recorder.Header().Get("Content-Type"))
		assertion.Contains(recorder.Body.String(), expected)
	}
}

func Test_Context_RenderWithStatusCoder(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)
	ctx.Logger = NewAppLogger("stderr", "")

	data := fakeContextStatusCoder{
		code: 123,
	}
	ctx.Render(NewDefaultRender(ctx.Response), data)

	assertion.Equal(123, ctx.Response.Status())
}

func Test_Context_RenderWithAbort(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)

	ctx := NewContext()
	ctx.Request = request
	ctx.Response.(*Response).reset(recorder)
	ctx.Logger = NewAppLogger("stderr", "")
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

	assertion.Equal("render", recorder.Body.String())
	assertion.EqualValues(-128, ctx.cursor)
}

func Test_Context_Next(t *testing.T) {
	assertion := assert.New(t)

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
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.middlewares = append(ctx.middlewares, middleware1, middleware2)
	ctx.Next()

	assertion.EqualValues(2, ctx.cursor)
	assertion.Equal(2, counter)
}

func Test_Context_Abort(t *testing.T) {
	assertion := assert.New(t)

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
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.middlewares = append(ctx.middlewares, middleware1, middleware2)
	ctx.Abort()
	ctx.Next()

	assertion.EqualValues(-128, ctx.cursor)
	assertion.Equal(0, counter)
}

func Benchmark_Context(b *testing.B) {
	w := httptest.NewRecorder()

	ctx := NewContext()
	ctx.Logger = NewAppLogger("null", "")

	for i := 0; i < b.N; i++ {
		ctx.run(w, nil, nil)
	}
}
