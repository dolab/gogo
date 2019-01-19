// +build !race

package gogo

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

var testServerWithTcpOnce sync.Once

func Test_ServerWithTcp(t *testing.T) {
	server := fakeTcpServer()
	server.GET("/server/run", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	var addr string
	testServerWithTcpOnce.Do(func() {
		go server.Run()
		for {
			if len(server.Address()) > 0 {
				break
			}
		}

		addr = server.Address()
	})

	client := httptesting.New(addr, false)

	request := client.New(t)
	request.Get("/server/run", nil)
	request.AssertStatus(http.StatusNotImplemented)
	request.AssertEmpty()
}

var benchmarkServerWithTcpOnce sync.Once

func Benchmark_ServerWithTcp(b *testing.B) {
	b.StopTimer()

	server := fakeServer()
	server.GET("/server/run", func(ctx *Context) {
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

		endpoint = "http://" + server.Address() + "/server/run"
	})

	client := &http.Client{}

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		client.Get(endpoint)
	}
}

var testServerWithUnixOnce sync.Once

func Test_ServerWithUnix(t *testing.T) {
	it := assert.New(t)

	server := fakeUnixServer()
	server.GET("/server/unix", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	var (
		client *http.Client
	)
	testServerWithUnixOnce.Do(func() {
		go server.Run()
		for {
			if len(server.Address()) > 0 {
				break
			}
		}

		unixAddr := server.Address()

		client = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
					return net.Dial("unix", unixAddr)
				},
			},
		}

	})
	defer server.localServ.Shutdown(context.Background())

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
	b.StopTimer()

	server := fakeUnixServer()
	server.GET("/server/unix", func(ctx *Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	var (
		client *http.Client
	)
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

	client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return unixConn, unixErr
			},
		},
	}

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		client.Get("http://unix/server/unix")
	}
}
