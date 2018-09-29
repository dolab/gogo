package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/golib/assert"
)

func Test_NewAppRoute(t *testing.T) {
	prefix := "/prefix"
	server := fakeServer()
	assertion := assert.New(t)

	route := NewAppRoute(prefix, server)
	assertion.Empty(route.middlewares)
	assertion.Equal(prefix, route.prefix)
	assertion.Equal(server, route.server)
}

func Test_RouteHandle(t *testing.T) {
	assertion := assert.New(t)
	server := fakeServer()

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
		server.Handle(method, testCase.path, testCase.handler)
	}

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// testing by http request
	for method, testCase := range testCases {
		request, _ := http.NewRequest(method, ts.URL+testCase.path, nil)

		response, err := http.DefaultClient.Do(request)
		assertion.Nil(err)

		body, err := ioutil.ReadAll(response.Body)
		response.Body.Close()

		switch method {
		case "HEAD":
			assertion.Empty(body)

		default:
			assertion.Equal(method, string(body))
		}
	}
}

func Test_RouteHandleWithTailSlash(t *testing.T) {
	assertion := assert.New(t)
	server := fakeServer()

	server.Handle("GET", "/:tailslash", func(ctx *Context) {
		ctx.Text("GET /:tailslash")
	})
	server.Handle("GET", "/:tailslash/*extraargs", func(ctx *Context) {
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
	assertion := assert.New(t)
	server := fakeServer()

	var n int32

	// proxied handler
	proxiedServer := server.Group("", func(ctx *Context) {
		atomic.AddInt32(&n, 1)

		ctx.Logger.Warn("proxy middleware")

		ctx.Next()
	})
	proxiedServer.Handle("GET", "/backend", func(ctx *Context) {
		atomic.AddInt32(&n, 1)

		ctx.Text("Proxied!")
	})

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			atomic.AddInt32(&n, 1)

			r.URL.Scheme = "http"
			r.URL.Host = r.Host
			r.URL.Path = "/backend"
			r.URL.RawPath = "/backend"
		},
	}
	server.ProxyHandle("GET", "/proxy", proxy, func(w Responser, b []byte) []byte {
		return []byte(strings.ToUpper(string(b)))
	})

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// testing by http request
	request, _ := http.NewRequest("GET", ts.URL+"/proxy", nil)

	response, err := http.DefaultClient.Do(request)
	assertion.Nil(err)
	assertion.EqualValues(3, n)

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	assertion.Nil(err)
	assertion.Equal("PROXIED!", string(body))
}

func Test_RouteMockHandle(t *testing.T) {
	assertion := assert.New(t)
	server := fakeServer()
	response := httptest.NewRecorder()

	// mock handler
	server.MockHandle("GET", "/mock", response, func(ctx *Context) {
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
	assertion := assert.New(t)
	server := fakeServer()
	route := server.Group("/group")

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

type testGroupController struct{}

func (t *testGroupController) Index(ctx *Context) {
	ctx.Text("GET /group")
}

func (t *testGroupController) Show(ctx *Context) {
	ctx.Text("GET /group/" + ctx.Params.Get("group"))
}

type testUserController struct{}

func (t *testUserController) ID() string {
	return "id"
}

func (t *testUserController) Show(ctx *Context) {
	ctx.Text("GET /group/" + ctx.Params.Get("group") + "/user/" + ctx.Params.Get("id"))
}

func Test_RouteResource(t *testing.T) {
	assertion := assert.New(t)
	server := fakeServer()

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// group resource
	group := server.Resource("group", &testGroupController{})

	// should work for GET /group/:group
	res, err := http.Get(ts.URL + "/group/my-group")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("GET /group/my-group", string(body))

	// user resource
	group.Resource("user", &testUserController{})

	// should work for GET /group/:group/user/:id
	res, err = http.Get(ts.URL + "/group/my-group/user/my-user")
	assertion.Nil(err)

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("GET /group/my-group/user/my-user", string(body))

	// error for not found
	res, err = http.Get(ts.URL + "/group/my-group/user/")
	assertion.Nil(err)

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal(http.StatusNotFound, res.StatusCode)
	assertion.Contains(string(body), "not found")
}

type testGroupMemberController struct{}

func (t *testGroupMemberController) Index(ctx *Context) {
	ctx.Text("GET /group/member")
}

func (t *testGroupMemberController) Show(ctx *Context) {
	ctx.Text("GET /group/member/" + ctx.Params.Get("member"))
}

func Test_RouteResourceWithSubPath(t *testing.T) {
	assertion := assert.New(t)
	server := fakeServer()

	// start server
	ts := httptest.NewServer(server)
	defer ts.Close()

	// member resource
	server.Resource("group/member", &testGroupMemberController{})

	// should work for GET /group/member/:group
	res, err := http.Get(ts.URL + "/group/member/my-group")
	assertion.Nil(err)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal("GET /group/member/my-group", string(body))

	// error for not found
	res, err = http.Get(ts.URL + "/group/member/my-group/user")
	assertion.Nil(err)

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	assertion.Equal(http.StatusNotFound, res.StatusCode)
	assertion.Contains(string(body), "not found")
}
