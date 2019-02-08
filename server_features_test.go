package gogo

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dolab/gogo/pkgs/hooks"
	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

func Test_ServerWithHealthz(t *testing.T) {
	server := fakeHealthzServer()

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	client := httptesting.New(server.Address(), false)

	request := client.New(t)
	request.Get(GogoHealthz, nil)
	request.AssertOK()
	request.AssertEmpty()
}

func Test_ServerWithThroughput(t *testing.T) {
	it := assert.New(t)
	logger := NewAppLogger("nil", "")
	config, _ := fakeConfig("application.throttle.yml")

	server := NewAppServer(config, logger)
	server.RequestReceived.PushFrontNamed(hooks.NewServerThrottleHook(1))

	server.GET("/server/throughput", func(ctx *Context) {
		ctx.SetStatus(http.StatusNoContent)
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
		go func() {
			defer wg.Done()

			request := ts.New(t)
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
	it.Contains(s, "I'm a teapot")

	// it should work for now
	request := ts.New(t)
	request.Get("/server/throughput", nil)
	request.AssertStatus(http.StatusNoContent)
	request.AssertEmpty()
}

func Test_ServerWithConcurrency(t *testing.T) {
	it := assert.New(t)
	logger := NewAppLogger("nil", "")
	config, _ := fakeConfig("application.throttle.yml")

	server := NewAppServer(config, logger)
	server.RequestReceived.PushFrontNamed(hooks.NewServerDemotionHook(1, 1))

	server.GET("/server/concurrency", func(ctx *Context) {
		if !ctx.Params.HasQuery("fast") {
			time.Sleep(time.Second)
		}

		ctx.SetStatus(http.StatusNoContent)
	})

	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	var (
		wg sync.WaitGroup

		routines = 3
	)

	bufc := make(chan string, routines)

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func(routine int) {
			defer wg.Done()

			request := ts.New(t)
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
	it.Contains(s, "Too Many Requests")

	// it should work now
	params := url.Values{}
	params.Add("fast", "")

	request := ts.New(t)
	request.Get("/server/concurrency", params)
	request.AssertStatus(http.StatusNoContent)
	request.AssertNotContains("Too Many Requests")
}

func Test_ServerWithTcp(t *testing.T) {
	server := fakeTcpServer()
	server.GET("/server/tcp", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	client := httptesting.New(server.Address(), false)

	request := client.New(t)
	request.Get("/server/tcp", nil)
	request.AssertStatus(http.StatusNotImplemented)
	request.AssertEmpty()
}

var benchmarkServerWithTCPOnce sync.Once

func Benchmark_ServerWithTCP(b *testing.B) {
	server := fakeServer()
	server.GET("/bench/tcp", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	var (
		endpoint string
	)
	benchmarkServerWithTCPOnce.Do(func() {
		go server.Run()
		for {
			if len(server.Address()) > 0 {
				break
			}
		}

		endpoint = "http://" + server.Address() + "/bench/tcp"
	})

	client := &http.Client{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(endpoint)
	}
}

func Test_ServerWithUnix(t *testing.T) {
	it := assert.New(t)

	server := fakeUnixServer()
	server.GET("/server/unix", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}
	defer os.Remove("/tmp/gogo.sock")

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return net.Dial("unix", server.Address())
			},
		},
	}

	// it should work
	response, err := client.Get("http://unix/server/unix")
	if it.Nil(err) {
		defer response.Body.Close()

		it.Equal(http.StatusNotImplemented, response.StatusCode)

		data, err := ioutil.ReadAll(response.Body)
		it.Nil(err)
		it.Empty(data)
	}

	// it should return error
	response, err = http.DefaultClient.Get("http://unix/server/unix")
	it.NotNil(err)
	it.Nil(response)
}

var benchmarkServerWithUnix sync.Once

func Benchmark_ServerWithUnix(b *testing.B) {
	server := fakeUnixServer()
	server.GET("/bench/unix", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	benchmarkServerWithUnix.Do(func() {
		go server.Run()
		for {
			if len(server.Address()) > 0 {
				break
			}
		}
	})
	defer os.Remove("/tmp/gogo.sock")

	unixConn, unixErr := net.Dial("unix", "/tmp/gogo.sock")
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return unixConn, unixErr
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get("http://unix/server/unix")
	}
}
