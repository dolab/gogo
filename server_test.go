package gogo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
	"github.com/golib/httprouter"
)

var (
	newMockServer = func() *AppServer {
		config, _ := newMockConfig("application.json")
		logger := NewAppLogger(config.Section().Logger.Output, "")

		return NewAppServer("test", config, logger)
	}
)

func Test_NewAppServer(t *testing.T) {
	assertion := assert.New(t)

	server := newMockServer()
	assertion.Implements((*http.Handler)(nil), server)
	assertion.IsType(&Context{}, server.pool.Get())
}

func Test_ServerNew(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	params := NewAppParams(request, httprouter.Params{})
	assertion := assert.New(t)

	server := newMockServer()
	ctx := server.new(recorder, request, params, nil)
	assertion.Equal(request, ctx.Request)
	assertion.Equal(recorder.Header().Get(server.requestId), ctx.Response.Header().Get(server.requestId))
	assertion.Equal(params, ctx.Params)
	assertion.Nil(ctx.settings)
	assertion.Nil(ctx.frozenSettings)
	assertion.Empty(ctx.handlers)
	assertion.EqualValues(-1, ctx.index)

	// creation
	newCtx := server.new(recorder, request, params, nil)
	assertion.NotEqual(fmt.Sprintf("%p", ctx), fmt.Sprintf("%p", newCtx))
}

func Test_ServerReuse(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	params := NewAppParams(request, httprouter.Params{})
	assertion := assert.New(t)

	server := newMockServer()
	ctx := server.new(recorder, request, params, nil)
	server.reuse(ctx)

	newCtx := server.new(recorder, request, params, nil)
	assertion.Equal(fmt.Sprintf("%p", ctx), fmt.Sprintf("%p", newCtx))
}

// // TODO: implements this later!
// func Test_ServerWithRequestTimeout(t *testing.T) {
// 	assertion := assert.New(t)

// 	config, _ := newMockConfig("request_timeout.json")
// 	server := NewAppServer(config, Logger)
// 	server.PUT("/server/request_timeout", func(ctx *Context) {
// 		b, err := ioutil.ReadAll(ctx.Request.Body)
// 		ctx.Request.Body.Close()

// 		assertion.Nil(err)
// 		assertion.Empty(string(b))

// 		ctx.Text("REQUEST TIMEOUT")
// 	})

// 	ts := httptest.NewAppServer(server)

// 	request, _ := http.NewRequest("PUT", ts.URL+"/server/request_timeout", bytes.NewBufferString("Ping!"))

// 	client := NewAppTesting(ts.URL, false)
// 	client.NewFilterRequest(t, request, func(r *http.Request) error {
// 		// wait 1s
// 		time.Sleep(1 * time.Second)

// 		return nil
// 	})
// 	client.AssertOK()
// 	client.AssertEmpty()
// }

func Test_ServerWithDisptacher(t *testing.T) {
	var dispatcher Dispatcher = func(r *http.Request) {
		r.URL.Path = "/server/dispatcher"
	}

	config, _ := newMockConfig("application.json")
	server := NewAppServer("test", config, NewAppLogger("stderr", ""))
	server.dispatcher = &dispatcher
	server.PUT("/server/dispatcher", func(ctx *Context) {
		ctx.Text("DISPATCHED")
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	request, _ := http.NewRequest("PUT", ts.URL+"/server/rehctapsid", bytes.NewBufferString("Ping!"))

	client := httptesting.New(ts.URL, false)
	client.NewRequest(t, request)
	client.AssertOK()
	client.AssertContains("DISPATCHED")
}
