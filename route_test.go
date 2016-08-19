package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"
	"strings"
	"testing"

	"github.com/golib/assert"
)

func Test_NewAppRoute(t *testing.T) {
	prefix := "/prefix"
	server := newMockServer()
	assertion := assert.New(t)

	route := NewAppRoute(prefix, server)
	assertion.Empty(route.Handlers)
	assertion.Equal(prefix, route.prefix)
	assertion.Equal(server, route.server)
}

func Test_RouteHandle(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server)
	assertion := assert.New(t)

	testCases := map[string]struct {
		path    string
		handler Middleware
	}{
		"PUT": {
			"/put",
			func(ctx *Context) {
				ctx.Text("PUT")
			},
		},
		"POST": {
			"/post",
			func(ctx *Context) {
				ctx.Text("POST")
			},
		},
		"GET": {
			"/get",
			func(ctx *Context) {
				ctx.Text("GET")
			},
		},
		"PATCH": {
			"/patch",
			func(ctx *Context) {
				ctx.Text("PATCH")
			},
		},
		"DELETE": {
			"/delete",
			func(ctx *Context) {
				ctx.Text("DELETE")
			},
		},
		"HEAD": {
			"/head",
			func(ctx *Context) {
				ctx.Text("HEAD")
			},
		},
		"OPTIONS": {
			"/options",
			func(ctx *Context) {
				ctx.Text("OPTIONS")
			},
		},
	}

	// register handlers
	for method, testCase := range testCases {
		route.Handle(method, testCase.path, testCase.handler)
	}

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// testing by http request
	for method, testCase := range testCases {
		request, _ := http.NewRequest(method, ts.URL+testCase.path, nil)

		res, err := http.DefaultClient.Do(request)
		assertion.Nil(err)

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()

		switch method {
		case "HEAD":
			assertion.Empty(body)

		default:
			assertion.Equal(method, string(body))
		}
	}
}

func Test_RouteHandleWithTailSlash(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server)
	assertion := assert.New(t)

	route.Handle("GET", "/:tailslash", func(ctx *Context) {
		ctx.Text("GET /:tailslash")
	})

	route.Handle("GET", "/:tailslash/*extraargs", func(ctx *Context) {
		ctx.Text("GET /:tailslash/*extraargs")
	})

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// without tail slash
	request, _ := http.NewRequest("GET", ts.URL+"/tailslash", nil)

	response, err := http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	assertion.Equal("GET /:tailslash", string(body))

	// with tail slash
	request, _ = http.NewRequest("GET", ts.URL+"/tailslash/?query", nil)

	response, err = http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err = ioutil.ReadAll(response.Body)
	response.Body.Close()

	assertion.Equal("GET /:tailslash", string(body))

	// with extra args without tail slash
	request, _ = http.NewRequest("GET", ts.URL+"/tailslash/extraargs", nil)

	response, err = http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err = ioutil.ReadAll(response.Body)
	response.Body.Close()

	assertion.Equal("GET /:tailslash/*extraargs", string(body))

	// with extra args with tail slash
	request, _ = http.NewRequest("GET", ts.URL+"/tailslash/extraargs/", nil)

	response, err = http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err = ioutil.ReadAll(response.Body)
	response.Body.Close()

	assertion.Equal("GET /:tailslash/*extraargs", string(body))
}

func Test_RouteProxyHandle(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server)
	assertion := assert.New(t)

	// proxied handler
	route.Use(func(ctx *Context) {
		ctx.Logger.Warn("v ...interface{}")
	})
	route.Handle("GET", "/proxy", func(ctx *Context) {
		ctx.Text("Proxied!")
	})

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = r.Host
			r.URL.Path = "/proxy"
			r.URL.RawPath = "/proxy"
		},
	}
	route.ProxyHandle("GET", "/mock", proxy, func(w Responser, b []byte) []byte {
		s := strings.ToUpper(string(b))

		w.Header().Set("Content-Length", strconv.Itoa(len(s)*2))

		return []byte(s + s)
	})

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// testing by http request
	request, _ := http.NewRequest("GET", ts.URL+"/mock", nil)

	res, err := http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assertion.Nil(err)
	assertion.Equal("PROXIED!PROXIED!", string(body))
}

func Test_RouteMockHandle(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server)
	response := httptest.NewRecorder()
	assertion := assert.New(t)

	// mock handler
	route.MockHandle("GET", "/mock", response, func(ctx *Context) {
		ctx.Text("MOCK")
	})

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// testing by http request
	request, _ := http.NewRequest("GET", ts.URL+"/mock", nil)

	res, err := http.DefaultClient.Do(request)
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assertion.Nil(err)
	assertion.Empty(body)

	assertion.Equal(http.StatusOK, response.Code)
	assertion.Equal("MOCK", response.Body.String())
}

func Test_RouteGroup(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server).Group("/group")
	assertion := assert.New(t)

	// register handler
	// GET /group/:method
	route.Handle("GET", "/:method", func(ctx *Context) {
		ctx.Text(ctx.Params.Get("method"))
	})

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/group/testing")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("testing", string(body))
}

type TestGroupController struct{}

func (t *TestGroupController) Index(ctx *Context) {
	ctx.Text(ctx.Params.Get("id") + "all")
}

func (t *TestGroupController) Show(ctx *Context) {
	ctx.Text(ctx.Params.Get("id") + "show")
}

type TestUserController struct{}

func (t *TestUserController) Show(ctx *Context) {
	ctx.Text(ctx.Params.Get("id") + ":" + ctx.Params.Get("user") + "show")
}

func (t *TestUserController) Id() string {
	return "user"
}

func Test_ResourceController(t *testing.T) {
	server := newMockServer()
	route := NewAppRoute("/", server)
	group := route.Resource("group", &TestGroupController{})
	assertion := assert.New(t)

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// test show group /group/:id
	res, err := http.Get(ts.URL + "/group/6666")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("6666show", string(body))

	// test user show /group/:id/user/:user
	group.Resource("user", &TestUserController{})
	res, err = http.Get(ts.URL + "/group/6666/user/82828")
	assertion.Nil(err)

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("6666:82828show", string(body))

	// test not found
	res, err = http.Get(ts.URL + "/group/6666/user/")
	assertion.Nil(err)

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("404 page not found\n", string(body))
	assertion.Equal(http.StatusNotFound, res.StatusCode)
}
