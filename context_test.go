package gogo

import (
	"crypto"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

func Test_NewContext(t *testing.T) {
	assertion := assert.New(t)

	server := newMockServer()
	ctx := NewContext(server)
	assertion.Equal(server, ctx.Server)
	assertion.EqualValues(-1, ctx.index)

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

func Test_Context_Next(t *testing.T) {
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
	ctx.handlers = append(ctx.handlers, middleware1, middleware2)
	ctx.Next()

	assertion.EqualValues(4, ctx.index)
	assertion.Equal(3, counter)
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
	ctx.handlers = append(ctx.handlers, middleware1, middleware2)
	ctx.Abort()
	ctx.Next()

	assertion.EqualValues(64, ctx.index)
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
