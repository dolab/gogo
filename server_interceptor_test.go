package gogo

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dolab/gogo/pkgs/interceptors"
	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

type stubInterceptor struct {
	name  string
	apply interceptors.Interceptor
}

func (stub *stubInterceptor) Name() string {
	return stub.name
}

func (stub *stubInterceptor) Config() []byte {
	return nil
}

func (stub *stubInterceptor) Priority() int {
	return 0
}

func (stub *stubInterceptor) Register(config interceptors.Configer) (interceptors.Interceptor, error) {
	return stub.apply, nil
}

func (stub *stubInterceptor) Reload(config interceptors.Configer) error {
	return nil
}

func (stub *stubInterceptor) Shutdown() error {
	return nil
}

type testService struct {
	triggered int64
	counter   int64
	v1        Grouper
}

func (svc *testService) Init(config Configer, group Grouper) {
	svc.triggered = 0
	svc.v1 = group
}

func (svc *testService) Middlewares() {
	svc.v1.Use(func(ctx *Context) {
		atomic.AddInt64(&svc.counter, 1)

		ctx.Next()
	})
}

func (svc *testService) Resources() {
	svc.v1.GET("/server/service", func(ctx *Context) {
		ctx.AddHeader("x-gogo-interceptor", strings.Join(ctx.Request.Header["X-Gogo-Interceptor"], ","))

		ctx.Text("Hello, service!")
	})
}

func (svc *testService) RequestReceived() []interceptors.Interface {
	return []interceptors.Interface{
		&stubInterceptor{
			name: "request_receved@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-interceptor", "Received")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) RequestRouted() []interceptors.Interface {
	return []interceptors.Interface{
		&stubInterceptor{
			name: "request_routed@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-interceptor", "Routed")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) ResponseReady() []interceptors.Interface {
	return []interceptors.Interface{
		&stubInterceptor{
			name: "response_ready@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-interceptor", "Ready")

				atomic.AddInt64(&svc.triggered, 1)
				return true
			},
		},
	}
}

func (svc *testService) ResponseAlways() []interceptors.Interface {
	return []interceptors.Interface{
		&stubInterceptor{
			name: "response_always@testing",
			apply: func(w http.ResponseWriter, r *http.Request) bool {
				r.Header.Add("x-gogo-interceptor", "Always")

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
	request.AssertHeader("x-gogo-interceptor", "Received,Routed")
	request.AssertContains("Hello, service!")

	it.EqualValues(1, service.counter)
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
			request.AssertHeader("x-gogo-interceptor", "Received,Routed")
			request.AssertContains("Hello, service!")
		}()
	}
	wg.Wait()

	it.EqualValues(1*max, service.counter)
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
