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

func Test_NewContext(t *testing.T) {
	assertion := assert.New(t)

	server := newMockServer()
	ctx := NewContext(server)
	assertion.Equal(server, ctx.Server)
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
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	request.Header.Add("X-Canonical-Key", "Canonical-Value")
	request.Header["x-normal-key"] = []string{"normal value"}
	params := NewAppParams(request, httprouter.Params{})
	assertion := assert.New(t)

	server := newMockServer()
	ctx := server.newContext(recorder, request, params, nil)
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
	counter := 0
	middleware1 := func(ctx *Context) {
		counter += 1

		ctx.Next()
	}
	middleware2 := func(ctx *Context) {
		counter += 1

		ctx.Next()
	}
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.handlers = append(ctx.handlers, middleware1, middleware2)
	ctx.Next()

	assertion.EqualValues(2, ctx.cursor)
	assertion.Equal(2, counter)
}

func Test_Context_Abort(t *testing.T) {
	counter := 0
	middleware1 := func(ctx *Context) {
		counter += 1

		ctx.Next()
	}
	middleware2 := func(ctx *Context) {
		counter += 2

		ctx.Next()
	}
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.handlers = append(ctx.handlers, middleware1, middleware2)
	ctx.Abort()
	ctx.Next()

	assertion.EqualValues(64, ctx.cursor)
	assertion.Equal(0, counter)
}

func Test_Context_Redirect(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
	ctx.Redirect(location)

	assertion.Equal(location, recorder.Header().Get("Location"))
}

func Test_Context_RedirectWithAbort(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	location := "https://www.example.com"
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.handlers = []Middleware{
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

func Test_Context_Render(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}

	testCases := map[string]struct {
		w    Render
		data interface{}
	}{
		"default render": {NewDefaultRender(ctx.Response), "default render"},
		"hashed render":  {NewHashRender(ctx.Response, crypto.MD5), "hashed render"},
		"text render":    {NewTextRender(ctx.Response), "text render"},
		`{"success":false,"error":"not found"}`: {NewJsonRender(ctx.Response), struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}{false, "not found"}},
		`js_callback({"success":true,"data":123});`: {NewJsonpRender(ctx.Response, "js_callback"), struct {
			Success bool `json:"success"`
			Data    int  `json:"data"`
		}{true, 123}},
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Result><Success>true</Success><Data>123</Data></Result>": {NewXmlRender(ctx.Response), struct {
			XMLName xml.Name `xml:"Result"`
			Success bool     `xml:"Success"`
			Data    int      `xml:"Data"`
		}{
			Success: true,
			Data:    123,
		}},
	}
	for expected, testCase := range testCases {
		recorder.Body.Reset()

		err := ctx.Render(testCase.w, testCase.data)
		assertion.Nil(err)
		assertion.Equal(expected, recorder.Body.String())
	}
}

func Test_Context_RenderWithAbort(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	assertion := assert.New(t)

	ctx := NewContext(newMockServer())
	ctx.Request = request
	ctx.Response = &Response{
		ResponseWriter: recorder,
	}
	ctx.Logger = NewAppLogger("stderr", "")
	ctx.handlers = []Middleware{
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
}
