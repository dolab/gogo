package gogo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

var (
	fakeApp = func(mode string) *AppServer {
		root, _ := os.Getwd()

		return New(mode, path.Join(root, "testdata"))
	}
)

func Test_New(t *testing.T) {
	app := fakeApp("test")
	app.GET("/gogo", func(ctx *Context) {
		ctx.Text("Hello, gogo!")
	})
	app.PUT("/ping", func(ctx *Context) {
		ctx.Text("pong")
	})

	ts := httptesting.NewServer(app, false)
	defer ts.Close()

	// GET
	request := ts.New(t)
	request.Get("/gogo")
	request.AssertOK()
	request.AssertContains("Hello, gogo!")

	// PUT
	request = ts.New(t)
	request.PutJSON("/ping", nil)
	request.AssertOK()
	request.AssertContains("pong")
}

func Test_NewWithModeConfig(t *testing.T) {
	it := assert.New(t)

	app := fakeApp("development")
	it.Equal("gogo for development", app.config.RunName())
}

func Test_NewWithFilename(t *testing.T) {
	it := assert.New(t)

	root, _ := os.Getwd()
	filename := path.Join(root, "testdata", "config", "application.yml")

	app := New("development", filename)
	it.Equal("gogo", app.config.RunName())
}

func Test_Gogo_Response(t *testing.T) {
	app := fakeApp("product")
	app.GET("/ping", func(ctx *Context) {
		ctx.SetStatus(http.StatusConflict)
		ctx.Return()
	})
	app.GET("/ghost", func(ctx *Context) {
		ctx.SetStatus(http.StatusConflict)
	})

	ts := httptesting.NewServer(app, false)
	defer ts.Close()

	// it should work for ping-pong
	request := ts.New(t)
	request.Get("/ping", nil)
	request.AssertStatus(http.StatusConflict)
	request.AssertEmpty()

	// it should work for ghost
	request = ts.New(t)
	request.Get("/ghost", nil)
	request.AssertStatus(http.StatusConflict)
	request.AssertEmpty()
}

func Benchmark_Gogo(b *testing.B) {
	it := assert.New(b)

	recorder := httptest.NewRecorder()
	reader := []byte("Hello,world!")

	app := fakeApp("product")
	app.GET("/bench", func(ctx *Context) {
		recorder.WriteHeader(http.StatusOK)
		recorder.Write(reader)
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	request, _ := http.NewRequest(http.MethodGet, ts.URL+"/bench", nil)
	_, err := http.DefaultClient.Do(request)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(reader, recorder.Body.Bytes())
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		http.DefaultClient.Do(request)

		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(reader, recorder.Body.Bytes())
	}
}

func Benchmark_Go(b *testing.B) {
	it := assert.New(b)

	recorder := httptest.NewRecorder()
	reader := []byte("Hello,world!")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.WriteHeader(http.StatusOK)
		recorder.Write(reader)
	}))
	defer ts.Close()

	request, _ := http.NewRequest(http.MethodGet, ts.URL+"/bench", nil)
	_, err := http.DefaultClient.Do(request)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(reader, recorder.Body.Bytes())
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		http.DefaultClient.Do(request)

		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(reader, recorder.Body.Bytes())
	}
}
