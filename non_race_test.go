// +build !race

package gogo

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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

var benchmarkServerWithTcpOnce sync.Once

func Benchmark_ServerWithTcp(b *testing.B) {
	server := fakeServer()
	server.GET("/bench/tcp", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	var (
		endpoint string
	)
	benchmarkServerWithTcpOnce.Do(func() {
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

func Test_Server_loggerNewWithReuse(t *testing.T) {
	it := assert.New(t)
	logger := NewAppLogger("nil", "")
	config, _ := fakeConfig("application.throttle.json")

	server := NewAppServer(config, logger)

	// new with request id
	alog := server.loggerNew("di-tseuqer-x")
	if it.NotNil(alog) {
		it.Implements((*Logger)(nil), alog)

		it.Equal("di-tseuqer-x", alog.RequestID())
	}
	server.loggerReuse(alog)

	// it should work for the same id
	blog := server.loggerNew("di-tseuqer-x")
	if it.NotNil(blog) {
		it.Equal("di-tseuqer-x", blog.RequestID())

		it.Equal(fmt.Sprintf("%p", alog), fmt.Sprintf("%p", blog))
	}
	server.loggerReuse(blog)

	clog := server.loggerNew("x-request-id")
	if it.NotNil(blog) {
		it.Equal("x-request-id", clog.RequestID())

		it.Equal(fmt.Sprintf("%p", alog), fmt.Sprintf("%p", clog))
	}
}

type timeoutRoundTripper struct {
	host            string
	requestTimeout  time.Duration
	responseTimeout time.Duration
}

func (rt *timeoutRoundTripper) RoundTrip(r *http.Request) (rs *http.Response, err error) {
	// issuedAt := time.Now()
	// defer func() {
	// 	fmt.Printf("%s\n", time.Now().Sub(issuedAt))
	// }()

	var conn net.Conn

	conn, err = net.Dial("tcp", r.Host)
	if err != nil {
		return
	}

	_, err = conn.Write([]byte(fmt.Sprintf("GET /server/timeout HTTP/1.1\r\nHost: %s\r\nContent-Length: 0\r\nAccept: text/html\r\n", rt.host)))
	if err != nil {
		return
	}

	if rt.requestTimeout > 0 {
		time.Sleep(rt.requestTimeout)
	}

	if rt.responseTimeout > 0 {
		_, err = conn.Write([]byte("X-Response-Timeout: true\r\n"))
		if err != nil {
			return
		}
	}

	_, err = conn.Write([]byte("\r\n\"\""))
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		return
	}
	defer conn.Close()

	rs = &http.Response{}
	lines := strings.Split(string(data), "\r\n")

	// parse HTTP/1.1 200 OK
	if len(lines) > 0 {
		fields := strings.SplitN(lines[0], " ", 3)
		if len(fields) == 3 {
			rs.StatusCode, _ = strconv.Atoi(fields[1])
			rs.Status = fields[2]
		}
	}

	// parse headers
	header := http.Header{}
	hasBody := false
	if len(lines) > 1 {
		for i := 1; i < len(lines)-1; i++ {
			if len(lines[i]) == 0 && i+2 == len(lines) {
				hasBody = true
				break
			}

			fields := strings.SplitN(lines[i], ": ", 2)
			if len(fields) == 2 {
				header.Add(fields[0], fields[1])
			}
		}
	}
	rs.Header = header

	// parse body
	body := ""
	if hasBody {
		body = lines[len(lines)-1]
	}
	rs.ContentLength = int64(len(body))
	rs.Body = ioutil.NopCloser(strings.NewReader(body))

	return
}

func Test_ServerWithTimeout(t *testing.T) {
	it := assert.New(t)

	server := fakeTimeoutServer()
	server.GET("/server/timeout", func(ctx *Context) {
		ctx.Response.FlushHeader()

		if ctx.HasHeader("X-Response-Timeout") {
			time.Sleep(1 * time.Second)
		}

		_, err := ctx.Response.Write([]byte("TIMEOUT"))
		it.Nil(err)
	})

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	endpoint := "http://" + server.Address() + "/server/timeout"

	// it should work
	client := &http.Client{
		Transport: &timeoutRoundTripper{
			host: server.Address(),
		},
	}

	response, err := client.Get(endpoint)
	if it.Nil(err) {
		data, err := ioutil.ReadAll(response.Body)
		if it.Nil(err) {
			it.Equal("TIMEOUT", string(data))
		}
	}

	// it should return timeout for request timeout
	requestTimeoutClient := &http.Client{
		Transport: &timeoutRoundTripper{
			host:           server.Address(),
			requestTimeout: 1 * time.Second,
		},
	}

	response, err = requestTimeoutClient.Get(endpoint)
	if it.NotNil(err) {
		it.Nil(response)

		it.Contains(err.Error(), "read: connection reset by peer")
	}

	// it should return timeout for response timeout
	responseTimeoutClient := &http.Client{
		Transport: &timeoutRoundTripper{
			host:            server.Address(),
			responseTimeout: 1 * time.Second,
		},
	}

	response, err = responseTimeoutClient.Get(endpoint)
	if it.Nil(err) {
		it.Empty(response.StatusCode)
		it.Empty(response.Status)

		data, err := ioutil.ReadAll(response.Body)
		if it.Nil(err) {
			it.Empty(data)
		}
	}
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
