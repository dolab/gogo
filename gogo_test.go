package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golib/assert"
)

var (
	fakeApp = func(mode string) *AppServer {
		root, _ := os.Getwd()

		return New(mode, path.Join(root, "skeleton", "gogo"))
	}
)

func Test_New(t *testing.T) {
	it := assert.New(t)

	app := fakeApp("test")
	app.GET("/gogo", func(ctx *Context) {
		ctx.Text("Hello, gogo!")
	})
	app.PUT("/ping", func(ctx *Context) {
		ctx.Text("pong")
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	// GET
	response, err := http.Get(ts.URL + "/gogo")
	if it.Nil(err) {
		body, err := ioutil.ReadAll(response.Body)
		if it.Nil(err) {
			response.Body.Close()

			it.Equal("Hello, gogo!", string(body))
		}
	}

	// PUT
	request, _ := http.NewRequest(http.MethodPut, ts.URL+"/ping", nil)

	response, err = http.DefaultClient.Do(request)
	if it.Nil(err) {
		body, err := ioutil.ReadAll(response.Body)
		if it.Nil(err) {
			response.Body.Close()

			it.Equal("pong", string(body))
		}
	}
}

func Test_NewWithModeConfig(t *testing.T) {
	it := assert.New(t)

	app := fakeApp("development")
	it.Equal("for development", app.config.RunName())
}

func Test_NewWithFilename(t *testing.T) {
	it := assert.New(t)

	root, _ := os.Getwd()
	filename := path.Join(root, "skeleton", "gogo", "config", "application.json")

	app := New("development", filename)
	it.Equal("gogo", app.config.RunName())
}

func Benchmark_Gogo(b *testing.B) {
	it := assert.New(b)

	recorder := httptest.NewRecorder()
	reader := []byte("Hello,world!")

	app := fakeApp("development")
	app.GET("/bench", func(ctx *Context) {
		recorder.WriteHeader(http.StatusOK)
		recorder.Write(reader)
	})

	ts := httptest.NewServer(app)
	defer ts.Close()

	request, _ := http.NewRequest(http.MethodGet, ts.URL+"/bench", nil)
	response, err := http.DefaultClient.Do(request)
	if it.Nil(err) {
		it.Equal(http.StatusOK, response.StatusCode)

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
	response, err := http.DefaultClient.Do(request)
	if it.Nil(err) {
		it.Equal(http.StatusOK, response.StatusCode)

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
