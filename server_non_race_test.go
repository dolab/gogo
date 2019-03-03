// +build !race

package gogo

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golib/assert"
)

func Test_Server_loggerNewWithReuse(t *testing.T) {
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
	timeout := 1 * time.Second

	server := fakeTimeoutServer()
	server.GET("/server/timeout", func(ctx *Context) {
		ctx.Response.FlushHeader()

		if ctx.HasHeader("X-Response-Timeout") {
			time.Sleep(timeout)
		}

		ctx.Response.Write([]byte("TIMEOUT"))
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
			requestTimeout: timeout,
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
			responseTimeout: timeout,
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
