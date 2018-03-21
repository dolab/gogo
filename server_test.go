package gogo

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
	"github.com/dolab/httpdispatch"
)

var (
	newMockServer = func() *AppServer {
		config, _ := newMockConfig("application.json")
		logger := NewAppLogger("stdout", "")

		return NewAppServer(config, logger)
	}
)

func Test_NewAppServer(t *testing.T) {
	assertion := assert.New(t)

	server := newMockServer()
	assertion.Implements((*http.Handler)(nil), server)
	assertion.IsType(&Context{}, server.context.Get())
}

func Test_ServerNewContext(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	params := NewAppParams(request, httpdispatch.Params{})

	server := newMockServer()
	ctx := server.newContext(request, params)
	ctx.run(recorder, nil)

	assertion.Equal(request, ctx.Request)
	assertion.Equal(recorder.Header().Get(server.requestID), ctx.Response.Header().Get(server.requestID))
	assertion.Equal(params, ctx.Params)
	assertion.Nil(ctx.settings)
	assertion.Nil(ctx.frozenSettings)
	assertion.Empty(ctx.middlewares)
	assertion.EqualValues(0, ctx.cursor)

	// creation
	newCtx := server.newContext(request, params)
	newCtx.run(recorder, nil)

	assertion.NotEqual(fmt.Sprintf("%p", ctx), fmt.Sprintf("%p", newCtx))
}

func Test_ServerReuseContext(t *testing.T) {
	assertion := assert.New(t)
	request, _ := http.NewRequest("GET", "https://www.example.com/resource?key=url_value&test=url_true", nil)
	params := NewAppParams(request, httpdispatch.Params{})

	server := newMockServer()
	ctx := server.newContext(request, params)
	server.reuseContext(ctx)

	newCtx := server.newContext(request, params)
	assertion.Equal(fmt.Sprintf("%p", ctx), fmt.Sprintf("%p", newCtx))
}

func Test_Server(t *testing.T) {
	config, _ := newMockConfig("application.json")
	logger := NewAppLogger("stdout", "")

	server := NewAppServer(config, logger)

	server.GET("/server", func(ctx *Context) {
		ctx.SetStatus(http.StatusNoContent)
		ctx.Return()
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	request := httptesting.New(ts.URL, false).New(t)
	request.Get("/server", nil)
	request.AssertStatus(http.StatusNoContent)
	request.AssertEmpty()
}

func Benchmark_Server(b *testing.B) {
	config, _ := newMockConfig("application.json")
	logger := NewAppLogger("stdout", "")

	server := NewAppServer(config, logger)
	server.GET("/server/benchmark", func(ctx *Context) {
		ctx.SetStatus(http.StatusNoContent)
		ctx.Return()
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	// NOTE: there is 37 allocs/op for client
	request, _ := http.NewRequest("GET", ts.URL+"/server/benchmark", nil)
	for i := 0; i < b.N; i++ {
		resp, _ := http.DefaultClient.Do(request)
		resp.Body.Close()
	}
}

func Test_ServerWithReturn(t *testing.T) {
	server := newMockServer()
	server.GET("/return", func(ctx *Context) {
		if contentType := ctx.Params.Get("content-type"); contentType != "" {
			ctx.SetHeader("Content-Type", contentType)
		}

		data := struct {
			XMLName xml.Name `json:"-"`
			Name    string   `xml:"Name"`
			Age     int      `xml:"Age"`
		}{
			XMLName: xml.Name{
				Local: "Result",
			},
			Name: "gogo",
			Age:  5,
		}

		ctx.Return(data)
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	// default render
	request := httptesting.New(ts.URL, false).New(t)
	request.Get("/return", nil)
	request.AssertOK()
	request.AssertHeader("Content-Type", "text/plain; charset=utf-8")
	request.AssertContains(`{{ Result} gogo 5}`)

	// json render with request header of accept
	request = httptesting.New(ts.URL, false).New(t)
	request.WithHeader("Accept", "application/json, text/xml, */*; q=0.01")
	request.Get("/return", nil)
	request.AssertOK()
	request.AssertHeader("Content-Type", "application/json")
	request.AssertContains(`{"Name":"gogo","Age":5}`)

	// default render with request query of content-Type=application/json
	params := url.Values{}
	params.Add("content-type", "application/json")

	request = httptesting.New(ts.URL, false).New(t)
	request.Get("/return?"+params.Encode(), nil)
	request.AssertOK()
	request.AssertHeader("Content-Type", "application/json")
	request.AssertContains(`{"Name":"gogo","Age":5}`)

	// xml render with request header of accept
	request = httptesting.New(ts.URL, false).New(t)
	request.WithHeader("Accept", "appication/json, text/xml, */*; q=0.01")
	request.Get("/return", nil)
	request.AssertOK()
	request.AssertHeader("Content-Type", "text/xml")
	request.AssertContains("<Result><Name>gogo</Name><Age>5</Age></Result>")

	// default render with request query of content-Type=text/xml
	params = url.Values{}
	params.Add("content-type", "text/xml")

	request = httptesting.New(ts.URL, false).New(t)
	request.Get("/return?"+params.Encode(), nil)
	request.AssertOK()
	request.AssertHeader("Content-Type", "text/xml")
	request.AssertContains(`<Result><Name>gogo</Name><Age>5</Age></Result>`)
}

func Test_ServerWithNotFound(t *testing.T) {
	server := newMockServer()

	ts := httptest.NewServer(server)
	defer ts.Close()

	request := httptesting.New(ts.URL, false).New(t)
	request.Get("/not/found", nil)
	request.AssertNotFound()
	request.AssertContains("Route(GET /not/found) not found")
}

func Test_ServerWithThroughput(t *testing.T) {
	assertion := assert.New(t)
	config, _ := newMockConfig("application.throttle.json")
	logger := NewAppLogger("stdout", "")

	server := NewAppServer(config, logger)
	server.newThrottle(1)

	server.GET("/server/throughput", func(ctx *Context) {
		ctx.SetStatus(http.StatusNoContent)
		ctx.Return()
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	var (
		wg sync.WaitGroup

		routines = 3
	)

	bufc := make(chan []byte, routines)

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func() {
			defer wg.Done()

			request := httptesting.New(ts.URL, false).New(t)
			request.Get("/server/throughput", nil)

			bufc <- request.ResponseBody
		}()
	}
	wg.Wait()

	close(bufc)

	s := ""
	for buf := range bufc {
		s += string(buf)
	}

	assertion.Contains(s, "I'm a teapot")
}

func Test_ServerWithConcurrency(t *testing.T) {
	assertion := assert.New(t)
	config, _ := newMockConfig("application.throttle.json")
	logger := NewAppLogger("stdout", "")

	server := NewAppServer(config, logger)
	server.newSlowdown(1, 1)

	server.GET("/server/concurrency", func(ctx *Context) {
		time.Sleep(time.Second)

		ctx.SetStatus(http.StatusNoContent)
		ctx.Return()
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	var (
		wg sync.WaitGroup

		routines = 2
	)

	bufc := make(chan string, routines)

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func(routine int) {
			defer wg.Done()

			request := httptesting.New(ts.URL, false).New(t)
			request.Get("/server/concurrency", nil)

			bufc <- fmt.Sprintf("[routine@#%d] %s", routine, string(request.ResponseBody))
		}(i)
	}
	wg.Wait()

	close(bufc)

	s := ""
	for buf := range bufc {
		s += string(buf)
	}

	assertion.Contains(s, "Too Many Requests")
}
