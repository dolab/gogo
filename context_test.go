package gogo

import (
	"crypto"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
	"github.com/golib/httprouter"
)

type testContextStatusCoder struct {
	code int
}

func (statusCoder testContextStatusCoder) StatusCode() int {
	return statusCoder.code
}

func Test_NewContext(t *testing.T) {
	assertion := assert.New(t)

	ctx := NewContext()
	assertion.EqualValues(-1, ctx.cursor)

	// settings
	assertion.Empty(ctx.settings)
	value, ok := ctx.Get("unknownSetting")
	assertion.False(ok)
	assertion.Empty(value)

	ctx.Set("middlewareKey", "middlewareValue")
	assertion.Equal(1, len(ctx.settings))
	value, ok = ctx.Get("middlewareKey")
	assertion.True(ok)
	assertion.Equal("middlewareValue", value)

	// MustGet
	assertion.Equal("middlewareValue", ctx.MustGet("middlewareKey"))
	assertion.Panics(func() {
		ctx.MustGet("unknownSetting")
	})

	// final settings
	assertion.Empty(ctx.frozenSettings)
	value, ok = ctx.GetFinal("unknownFinalSetting")
	assertion.False(ok)
	assertion.Empty(value)

	err := ctx.SetFinal("middlewareFinalKey", "middlewareFinalValue")
	assertion.Nil(err)
	assertion.Equal(1, len(ctx.frozenSettings))
	value, ok = ctx.GetFinal("middlewareFinalKey")
	assertion.True(ok)
	assertion.Equal("middlewareFinalValue", value)

	err = ctx.SetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	assertion.EqualError(err, ErrSettingsKey.Error())
	assertion.Equal(1, len(ctx.frozenSettings))
	value, ok = ctx.GetFinal("middlewareFinalKey")
	assertion.True(ok)
	assertion.Equal("middlewareFinalValue", value)

	// MustSetFinal
	assertion.Panics(func() {
		ctx.MustSetFinal("middlewareFinalKey", "newMiddlewareFinalValue")
	})

	// MusetGetFinal
	assertion.Panics(func() {
		ctx.MustGetFinal("unknownMiddlewareFinalKey")
	})
}

func Test_Context_RequestHeader(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	request.Header.Add("X-Canonical-Key", "Canonical-Value")
	request.Header["x-normal-key"] = []string{"normal value"}

	params := NewAppParams(request, httprouter.Params{})

	server := newMockServer()
	ctx := server.newContext(request, params)
	ctx.run(recorder, nil)

	assertion.True(ctx.HasRawHeader("X-Canonical-Key"))
	assertion.False(ctx.HasRawHeader("x-canonical-key"))
	assertion.True(ctx.HasHeader("X-Canonical-Key"))
	assertion.True(ctx.HasHeader("x-canonical-key"))
	assertion.True(ctx.HasRawHeader("x-normal-key"))
	assertion.False(ctx.HasRawHeader("X-Normal-Key"))
	assertion.False(ctx.HasHeader("x-normal-key"))
	assertion.False(ctx.HasHeader("X-Normal-Key"))
	assertion.Equal("Canonical-Value", ctx.RawHeader("X-Canonical-Key"))
	assertion.Empty(ctx.RawHeader("x-canonical-key"))
	assertion.Equal("Canonical-Value", ctx.Header("X-Canonical-Key"))
	assertion.Equal("Canonical-Value", ctx.Header("x-canonical-key"))
	assertion.Equal("normal value", ctx.RawHeader("x-normal-key"))
	assertion.Empty(ctx.RawHeader("X-Normal-Key"))
	assertion.Empty(ctx.Header("x-normal-key"))
	assertion.Empty(ctx.Header("X-Normal-Key"))
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

	assertion.EqualValues(64, ctx.cursor)
	assertion.Equal(0, counter)
}

func Test_Context_Redirect(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
	ctx.Redirect(location)

	assertion.Equal(location, recorder.Header().Get("Location"))
	assertion.EqualValues(63, ctx.cursor)
}

func Test_Context_RedirectWithAbort(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"

	ctx := NewContext()
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
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
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}

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
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}

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
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}

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
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
	ctx.Logger = NewAppLogger("stderr", "")

	data := testContextStatusCoder{
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
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
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
	assertion.EqualValues(64, ctx.cursor)
}
