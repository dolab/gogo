// +build !race

package gogo

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

type testMiddleware struct {
	name  string
	apply func(w http.ResponseWriter, r *http.Request) bool
}

func (m *testMiddleware) Name() string {
	return m.name
}

func (m *testMiddleware) Config() []byte {
	return nil
}

func (m *testMiddleware) Priority() int {
	return 0
}

func (m *testMiddleware) Register(config MiddlewareConfiger) (func(w http.ResponseWriter, r *http.Request) bool, error) {
	return m.apply, nil
}

func (m *testMiddleware) Reload(config MiddlewareConfiger) error {
	return nil
}

func (m *testMiddleware) Shutdown() error {
	return nil
}

type testService struct {
	triggered   int64
	middlewares int64
	v1          Grouper
}

func (svc *testService) Init(config Configer, group Grouper) {
	svc.triggered = 0
	svc.v1 = group
}

func (svc *testService) Filters() {
	svc.v1.Use(func(ctx *Context) {
		atomic.AddInt64(&svc.middlewares, 1)

		ctx.Next()
	})
}

func (svc *testService) Resources() {
	svc.v1.GET("/server/service", func(ctx *Context) {
		ctx.AddHeader("x-gogo-middleware", strings.Join(ctx.Request.Header["X-Gogo-Middleware"], ","))

		ctx.Text("Hello, service!")
	})
}

func (svc *testService) RequestReceived() []Middlewarer {
	return []Middlewarer{
		&testMiddleware{
			name: "request_receved@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-middleware", "Received")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) RequestRouted() []Middlewarer {
	return []Middlewarer{
		&testMiddleware{
			name: "request_routed@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-middleware", "Routed")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) ResponseReady() []Middlewarer {
	return []Middlewarer{
		&testMiddleware{
			name: "response_ready@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-middleware", "Ready")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) ResponseAlways() []Middlewarer {
	return []Middlewarer{
		&testMiddleware{
			name: "response_always@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-middleware", "Always")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func Test_Server_NewService(t *testing.T) {
	it := assert.New(t)
	service := &testService{}
	server := fakeServer()
	server.NewService(service)

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	client := httptesting.New(server.Address(), false)

	request := client.New(t)
	request.Get("/server/service", nil)
	request.AssertOK()
	request.AssertHeader("x-gogo-middleware", "Received,Routed,Ready,Always")
	request.AssertContains("Hello, service!")

	it.EqualValues(1, service.middlewares)
	it.EqualValues(4, service.triggered)
}

func Test_Server_NewServiceWithConcurrency(t *testing.T) {
	it := assert.New(t)
	service := &testService{}
	server := fakeServer()
	server.NewService(service)

	go server.Run()
	for {
		if len(server.Address()) > 0 {
			break
		}
	}

	client := httptesting.New(server.Address(), false)

	var (
		max = 10

		wg sync.WaitGroup
	)

	wg.Add(max)
	for i := 0; i < max; i++ {
		go func() {
			defer wg.Done()

			request := client.New(t)
			request.Get("/server/service", nil)
			request.AssertOK()
			request.AssertHeader("x-gogo-middleware", "Received,Routed,Ready,Always")
			request.AssertContains("Hello, service!")
		}()
	}
	wg.Wait()

	it.EqualValues(1*max, service.middlewares)
	it.EqualValues(4*max, service.triggered)
}

var benchmarkServiceOnce sync.Once

func Benchmark_Server_Service(b *testing.B) {
	service := &testService{}
	server := fakeServer()
	server.NewService(service)

	var (
		endpoint string
	)
	benchmarkServiceOnce.Do(func() {
		go server.Run()
		for {
			if len(server.Address()) > 0 {
				break
			}
		}

		endpoint = "http://" + server.Address() + "/server/service"
	})

	client := &http.Client{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(endpoint)
	}
}
