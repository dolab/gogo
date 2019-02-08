package gogo

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

var (
	fakeServer = func() *AppServer {
		logger := NewAppLogger("nil", "")
		logger.SetSkip(3)

		config, _ := fakeConfig("application.yml")

		return NewAppServer(config, logger)
	}

	fakeTcpServer = func() *AppServer {
		logger := NewAppLogger("nil", "")
		logger.SetSkip(3)

		config, _ := fakeConfig("application.tcp.yml")

		return NewAppServer(config, logger)
	}

	fakeUnixServer = func() *AppServer {
		logger := NewAppLogger("nil", "")
		logger.SetSkip(3)

		config, _ := fakeConfig("application.unix.yml")

		return NewAppServer(config, logger)
	}

	fakeTimeoutServer = func() *AppServer {
		logger := NewAppLogger("nil", "")
		logger.SetSkip(3)

		config, _ := fakeConfig("application.timeout.yml")

		return NewAppServer(config, logger)
	}

	fakeHealthzServer = func() *AppServer {
		logger := NewAppLogger("nil", "")
		logger.SetSkip(3)

		config, _ := fakeConfig("application.healthz.yml")

		return NewAppServer(config, logger)
	}
)

func Test_NewAppServer(t *testing.T) {
	it := assert.New(t)

	server := fakeServer()
	it.Implements((*http.Handler)(nil), server)
}

func Test_Server(t *testing.T) {
	server := fakeServer()

	server.GET("/server", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
		ctx.Return()
	})

	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	request := ts.New(t)
	request.Get("/server", nil)
	request.AssertStatus(http.StatusNotImplemented)
	request.AssertEmpty()
}

func Benchmark_Server(b *testing.B) {
	it := assert.New(b)
	server := fakeServer()
	server.GET("/server/benchmark", func(ctx *Context) {
		ctx.SetStatus(http.StatusNoContent)
	})

	ts := httptest.NewServer(server)
	defer ts.Close()

	r, _ := http.NewRequest("GET", ts.URL+"/server/benchmark", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, r)
	it.Equal(http.StatusNoContent, w.Code)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.ServeHTTP(w, r)
	}
}

func Benchmark_ServerWithReader(b *testing.B) {
	it := assert.New(b)

	reader := []byte("Hello,world!")

	server := fakeServer()
	server.GET("/server/benchmark/reader", func(ctx *Context) {
		ctx.Return(bytes.NewReader(reader))
	})

	r, _ := http.NewRequest("GET", "/server/benchmark/reader", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, r)
	it.Equal(http.StatusOK, w.Code)
	it.Equal(reader, w.Body.Bytes())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.ServeHTTP(w, r)
	}
}

func Test_Server_Race(t *testing.T) {
	it := assert.New(t)
	logger := NewAppLogger("nil", "")
	config, _ := fakeConfig("application.yml")

	server := NewAppServer(config, logger)
	server.RequestReceived.PushFront(func(w http.ResponseWriter, r *http.Request) bool {
		time.Sleep(time.Millisecond)
		r.Header.Add("X-Server-Race", r.Header.Get("X-Server-Race"))
		return true
	})

	server.GET("/server/race", func(ctx *Context) {
		ctx.Return(ctx.Header("X-Server-Race"))
	})

	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	var (
		wg sync.WaitGroup

		routines = 3
	)

	bufc := make(chan []byte, routines)

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func(n int) {
			defer wg.Done()

			request := ts.New(t)
			request.WithHeader("X-Server-Race", "Race#"+strconv.Itoa(n))
			request.Get("/server/race", nil)

			bufc <- request.ResponseBody
		}(i)
	}
	wg.Wait()

	close(bufc)

	s := ""
	for buf := range bufc {
		s += string(buf)
	}

	for i := 0; i < routines; i++ {
		it.Equal(1, strings.Count(s, "Race#"+strconv.Itoa(i)))
	}
}

func Test_ServerWithReturn(t *testing.T) {
	server := fakeServer()
	server.GET("/server/return", func(ctx *Context) {
		if contentType := ctx.Params.Get("content-type"); contentType != "" {
			ctx.SetHeader("Content-Type", contentType)
		}

		data := struct {
			XMLName xml.Name `json:"-"`
			Name    string   `json:"Name"`
			Age     int      `json:"Age"`
		}{
			XMLName: xml.Name{
				Local: "Result",
			},
			Name: "gogo",
			Age:  5,
		}

		ctx.Return(data)
	})

	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// default render
	request := ts.New(t)
	request.Get("/server/return", nil)
	request.AssertStatus(http.StatusOK)
	request.AssertHeader("Content-Type", "text/plain; charset=utf-8")
	request.AssertContains(`{{ Result} gogo 5}`)

	// json render with request header of accept
	request = ts.New(t)
	request.WithHeader("Accept", "application/json, text/xml, */*; q=0.01")
	request.Get("/server/return", nil)
	request.AssertStatus(http.StatusOK)
	request.AssertHeader("Content-Type", "application/json")
	request.AssertContains(`{"Name":"gogo","Age":5}`)

	// default render with request query of content-Type=application/json
	params := url.Values{}
	params.Add("content-type", "application/json")

	request = ts.New(t)
	request.Get("/server/return?"+params.Encode(), nil)
	request.AssertStatus(http.StatusOK)
	request.AssertHeader("Content-Type", "application/json")
	request.AssertContains(`{"Name":"gogo","Age":5}`)

	// xml render with request header of accept
	request = ts.New(t)
	request.WithHeader("Accept", "appication/json, text/xml, */*; q=0.01")
	request.Get("/server/return", nil)
	request.AssertStatus(http.StatusOK)
	request.AssertHeader("Content-Type", "text/xml")
	request.AssertContains("<Result><Name>gogo</Name><Age>5</Age></Result>")

	// default render with request query of content-Type=text/xml
	params = url.Values{}
	params.Add("content-type", "text/xml")

	request = ts.New(t)
	request.Get("/server/return?"+params.Encode(), nil)
	request.AssertStatus(http.StatusOK)
	request.AssertHeader("Content-Type", "text/xml")
	request.AssertContains(`<Result><Name>gogo</Name><Age>5</Age></Result>`)
}

func Test_ServerWithNotFound(t *testing.T) {
	server := fakeServer()

	ts := httptest.NewServer(server)
	defer ts.Close()

	request := httptesting.New(ts.URL, false).New(t)
	request.Get("/server/not/found", nil)
	request.AssertNotFound()
	request.AssertContains("Request(GET /server/not/found): not found")
}

func Test_Server_Shutdown(t *testing.T) {
	it := assert.New(t)

	server := fakeTimeoutServer()
	server.GET("/server/shutdown", func(ctx *Context) {
		ctx.Text("SHUTDOWN")
	})

	go server.Serve()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	// it should not work any more
	server.Shutdown()

	endpoint := "http://" + server.Address() + "/server/shutdown"

	response, err := http.Get(endpoint)
	if it.NotNil(err) {
		it.Nil(response)

		it.Contains(err.Error(), "connect: connection refused")
	}
}

func Test_Server_ShutdownWithGhost(t *testing.T) {
	it := assert.New(t)

	server := fakeTimeoutServer()
	server.GET("/server/shutdown/ghost", func(ctx *Context) {
		ctx.Text("SHUTDOWN")
	})

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	// it should not work any more
	server.Shutdown()

	endpoint := "http://" + server.Address() + "/server/shutdown/ghost"

	response, err := http.Get(endpoint)
	if it.Nil(err) {
		defer response.Body.Close()

		data, err := ioutil.ReadAll(response.Body)
		if it.Nil(err) {
			it.Equal("SHUTDOWN", string(data))
		}
	}
}

// func Test_ServerWithMethodNotAllowed(t *testing.T) {
// 	server := fakeServer()
// 	server.HEAD("/server/method/not/allowed", func(ctx *Context) {
// 		ctx.Text("MethodNotAllowed")
// 	})

// 	ts := httptest.NewServer(server)
// 	defer ts.Close()

// 	request := httptesting.New(ts.URL, false).New(t)
// 	request.Get("/server/method/not/allowed", nil)
// 	request.AssertNotFound()
// 	request.AssertContains("Request(GET /server/method/not/allowed): method not allowed")
// }

func Test_Server_loggerNew(t *testing.T) {
	it := assert.New(t)
	logger := NewAppLogger("nil", "")
	config, _ := fakeConfig("application.throttle.yml")

	server := NewAppServer(config, logger)

	// new with request id
	alog := server.loggerNew("di-tseuqer-x")
	if it.NotNil(alog) {
		it.Implements((*Logger)(nil), alog)

		it.Equal("di-tseuqer-x", alog.RequestID())
	}

	blog := server.loggerNew("x-request-id")
	if it.NotNil(blog) {
		it.Implements((*Logger)(nil), blog)

		it.NotEqual(fmt.Sprintf("%p", alog), fmt.Sprintf("%p", blog))
	}

	it.Equal("di-tseuqer-x", alog.RequestID())
}
