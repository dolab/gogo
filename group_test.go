package gogo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dolab/httptesting"
	"github.com/golib/assert"
)

func Test_NewAppGroup(t *testing.T) {
	it := assert.New(t)
	server := fakeServer()
	prefix := "/prefix"

	group := NewAppGroup(prefix, server)
	it.Equal(server, group.server)
	it.Equal(prefix, group.prefix)
	it.NotNil(group.handler)
	it.Empty(group.filters)
}

func Test_Group_Proxy(t *testing.T) {
	it := assert.New(t)
	server := fakeServer()
	proxiedServer := fakeServer()

	var n int32

	// proxied handler
	proxiedServer.Handle("GET", "/backend", func(ctx *Context) {
		atomic.AddInt32(&n, 1)

		ctx.SetStatus(http.StatusOK)
		ctx.Text("I AM BACKEND!")
	})

	// start proxy server
	proxiedTs := httptest.NewServer(proxiedServer)
	defer proxiedTs.Close()

	proxiedURL, _ := url.Parse(proxiedTs.URL)

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			atomic.AddInt32(&n, 1)

			r.URL.Scheme = "http"
			r.URL.Host = proxiedURL.Host
			r.URL.Path = "/backend"
			r.URL.RawPath = "/backend"
		},
	}
	server.Proxy("GET", "/proxy", proxy, func(w Responser, b []byte) []byte {
		return []byte(strings.ToUpper(string(b)))
	})

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// testing by http request
	request := ts.New(t)
	request.Get("/proxy")
	request.AssertOK()
	request.AssertContains("I AM BACKEND!")

	it.EqualValues(2, n)
}

func Test_Group_Handle(t *testing.T) {
	it := assert.New(t)
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
		if it.Nil(err) {

			body, err := ioutil.ReadAll(response.Body)
			if it.Nil(err) {
				response.Body.Close()

				switch method {
				case "HEAD":
					it.Empty(body)

				default:
					it.Equal(method, string(body))
				}
			}
		}
	}
}

func Test_Group_HandleWithTailSlash(t *testing.T) {
	server := fakeServer()

	server.Handle("GET", "/:tailslash", func(ctx *Context) {
		ctx.Text("GET /:tailslash")
	})
	server.Handle("GET", "/:tailslash/*extraargs", func(ctx *Context) {
		ctx.Text("GET /:tailslash/*extraargs")
	})

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// without tail slash
	request := ts.New(t)
	request.Get("/tailslash")
	request.AssertOK()
	request.AssertContains("GET /:tailslash")

	// with tail slash
	request = ts.New(t)
	request.Get("/tailslash/?query")
	request.AssertOK()
	request.AssertContains("GET /:tailslash")

	// with extra args without tail slash
	request = ts.New(t)
	request.Get("/tailslash/extraargs")
	request.AssertOK()
	request.AssertContains("GET /:tailslash/*extraargs")

	// with extra args with tail slash
	request = ts.New(t)
	request.Get("/tailslash/extraargs/")
	request.AssertOK()
	request.AssertContains("GET /:tailslash/*extraargs")
}

func Test_Group_MockHandle(t *testing.T) {
	it := assert.New(t)
	server := fakeServer()
	recorder := httptest.NewRecorder()

	// mock handler
	server.MockHandle("GET", "/mock", recorder, func(ctx *Context) {
		ctx.Text("MOCK")
	})

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// testing by http request
	request := ts.New(t)
	request.Get("/mock")
	request.AssertOK()
	request.AssertEmpty()

	it.Equal(http.StatusOK, recorder.Code)
	it.Equal("MOCK", recorder.Body.String())
}

func Test_Group_NewGroup(t *testing.T) {
	server := fakeServer()
	group := server.NewGroup("/group")

	// register handler
	// GET /group/:method
	group.Handle("GET", "/:method", func(ctx *Context) {
		ctx.Text(ctx.Params.Get("method"))
	})

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	request := ts.New(t)
	request.Get("/group/testing")
	request.AssertOK()
	request.AssertContains("testing")
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

func Test_Group_Resource(t *testing.T) {
	server := fakeServer()

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// group resource
	group := server.Resource("group", &testGroupController{})

	// should work for GET /group/:group
	request := ts.New(t)
	request.Get("/group/my-group")
	request.AssertOK()
	request.AssertContains("GET /group/my-group")

	// user resource
	group.Resource("user", &testUserController{})

	// should work for GET /group/:group/user/:id
	request = ts.New(t)
	request.Get("/group/my-group/user/my-user")
	request.AssertOK()
	request.AssertContains("GET /group/my-group/user/my-user")

	// error for not found
	request = ts.New(t)
	request.Get("/group/my-group/user/")
	request.AssertStatus(http.StatusNotFound)
	request.AssertContains("not found")
}

type testGroupMemberController struct{}

func (t *testGroupMemberController) Index(ctx *Context) {
	ctx.Text("GET /group/member")
}

func (t *testGroupMemberController) Show(ctx *Context) {
	ctx.Text("GET /group/member/" + ctx.Params.Get("member"))
}

func Test_Group_ResourceWithSubPath(t *testing.T) {
	server := fakeServer()

	// start server
	ts := httptesting.NewServer(server, false)
	defer ts.Close()

	// member resource
	server.Resource("group/member", &testGroupMemberController{})

	// should work for GET /group/member/:group
	request := ts.New(t)
	request.Get("/group/member/my-group")
	request.AssertOK()
	request.AssertContains("GET /group/member/my-group")

	// error for not found
	request = ts.New(t)
	request.Get("/group/member/my-group/user")
	request.AssertStatus(http.StatusNotFound)
	request.AssertContains("not found")
}

func Test_GroupWithReservedRoute(t *testing.T) {
	it := assert.New(t)
	server := fakeServer()

	it.Panics(func() {
		server.GET("/-/healthz", func(ctx *Context) {
			ctx.SetStatus(http.StatusInternalServerError)
		})
	})

	it.Panics(func() {
		server.Handler(http.MethodGet, "/-/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
	})

	it.Panics(func() {
		server.Handle(http.MethodGet, "/-/healthz", func(ctx *Context) {
			ctx.SetStatus(http.StatusInternalServerError)
		})
	})

	it.Panics(func() {
		server.MockHandle(http.MethodGet, "/-/healthz", httptest.NewRecorder(), func(ctx *Context) {
			ctx.SetStatus(http.StatusInternalServerError)
		})
	})
}
