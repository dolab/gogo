# gogo

[![Build Status](https://travis-ci.org/dolab/gogo.svg?branch=master&style=flat)](https://travis-ci.org/dolab/gogo) [![Coverage](http://gocover.io/_badge/github.com/dolab/gogo?0)](http://gocover.io/github.com/dolab/gogo) [![GoDoc](https://godoc.org/github.com/dolab/gogo?status.svg)](http://godoc.org/github.com/dolab/gogo)

`gogo` is an open source, high performance RESTful api framework for the [Golang](https://golang.org) programming language.

It's heavily inspired from [rails](http://rubyonrails.org/) and [neko](https://github.com/rocwong/neko).

## Installation

```bash
$ go get github.com/dolab/gogo
```

- Create application using scaffold tools

```bash
$ go get github.com/dolab/gogo/gogo

# show gogo helps
$ gogo -h

# create a new application
$ gogo new myapp

# fix application import path
$ cd myapp
$ source env.sh

# run development server
$ make godev

# run test
$ make
```


## Getting Started

- Normal

```go
package main

import (
	"net/http"

	"github.com/dolab/gogo"
)

func main() {
	app := gogo.New("development", "/path/to/your/config")

	// GET /
	app.GET("/", func(ctx *gogo.Context) {
		ctx.Text("Hello, gogo!")
	})

	// GET /hello/:name
	app.HandlerFunc(http.MethodGet, "/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		params := gogo.NewAppParams(r)

		name := params.Get("name")
		if name == "" {
			name = "gogo"
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, " + name + "!"))
	})

	app.Run()
}
```

- Using middlewares

```go
package main

import (
    "github.com/dolab/gogo"
)

func main() {
    app := gogo.New("development", "/path/to/your/config")

    // avoid server quit by registering a recovery middleware
    app.Use(func(ctx *gogo.Context) {
        if panicErr := recover(); panicErr != nil {
            ctx.Abort()

            ctx.Logger.Errorf("[PANICED] %v", panicErr)
            return
        }

        ctx.Next()
    })

    // GET /
    app.GET("/", func(ctx *gogo.Context) {
        panic("Oops ~ ~ ~")
    })

    app.Run()
}
```

- Using group

```go
package main

import (
    "encoding/base64"
    "net/http"
    "strings"

    "github.com/dolab/gogo"
)

func main() {
    app := gogo.New("development", "/path/to/your/config")

    // avoid server quit by registering recovery func global
    app.Use(func(ctx *gogo.Context) {
        if panicErr := recover(); panicErr != nil {
            ctx.Abort()

            ctx.Logger.Errorf("[PANICED] %v", panicErr)
        }

        ctx.Next()
    })

    // GET /
    app.GET("/", func(ctx *gogo.Context) {
        panic("Oops ~ ~ ~")
    })

    // prefix resources with /v1 and apply basic auth middleware for all sub-resources
    // NOTE: it combines recovery middleware from previous.
    v1 := app.Group("/v1", func(ctx *gogo.Context) {
        auth := ctx.Header("Authorization")
        if !strings.HasPrefix(auth, "Basic ") {
            ctx.Abort()

            ctx.SetStatus(http.StatusForbidden)
            ctx.Return()
            return
        }

        b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
        if err != nil {
            ctx.Logger.Errorf("Base64 decode error: %v", err)
            ctx.Abort()

            ctx.SetStatus(http.StatusForbidden)
            ctx.Return()
            return
        }

        tmp := strings.SplitN(string(b), ":", 2)
        if len(tmp) != 2 || tmp[0] != "gogo" || tmp[1] != "ogog" {
            ctx.Abort()

            ctx.SetStatus(http.StatusForbidden)
            ctx.Return()
            return
        }

        // settings which can used by following middlewares and handler
        ctx.Set("username", tmp[0])

        ctx.Next()
    })

    // GET /v1
    v1.GET("/", func(ctx *gogo.Context) {
        username := ctx.MustGet("username").(string)

        ctx.Text("Hello, " + username + "!")
    })

    app.Run()
}
```

- Use Resource Controller

You can implement a *controller* with optional `Index`, `Create`, `Explore`, `Show`, `Update` and `Destroy` methods, 
and use `app.Resource("resourceName", &MyController)` to register all RESTful routes auto.

NOTE: When your resource has a inheritance relationship, there **MUST NOT** be two same id key.
you can overwrite default id key by implementing `ControllerID` interface.

```go
package main

import (
    "github.com/dolab/gogo"
)

type GroupController struct{}

// GET /group
func (t *GroupController) Index(ctx *gogo.Context) {
	ctx.Text("GET /group")
}

// GET /group/:group
func (t *GroupController) Show(ctx *gogo.Context) {
	ctx.Text("GET /group/" + ctx.Params.Get("group"))
}

type UserController struct{}

// overwrite default :user key with :id
func (t *UserController) ID() string {
	return "id"
}

// GET /group/:group/user/:id
func (t *UserController) Show(ctx *gogo.Context) {
	ctx.Text("GET /group/" + ctx.Params.Get("group") + "/user/" + ctx.Params.Get("id"))
}

func main() {
	app := gogo.New("development", "/path/to/your/config")

	// register group controller with default :group key
	group := app.Resource("group", &GroupController{})

	// nested user controller within group resource
	// NOTE: it overwrites default :user key by implmenting ControllerID interface.
	group.Resource("user", &UserController{})

	app.Run()
}
```

## Configures

- Server

```json
{
    "addr": "localhost",
    "port": 9090,
    "throttle": 3000, // RPS, throughput of per-seconds
    "slowdown": 30000, // TPS, concurrency of server
    "request_timeout": 30,
    "response_timeout": 30,
    "ssl": false,
    "ssl_cert": "/path/to/ssl/cert",
    "ssl_key": "/path/to/ssl/key",
    "request_id": "X-Request-Id"
}
```

## TODOs

- [x] server config context
- [x] support http.Request context
- [ ] scoffold && generator
- [ ] mountable third-part app

## Thanks

- [httprouter](https://github.com/julienschmidt/httprouter)

## Author

[Spring MC](https://twitter.com/mcspring)

## LICENSE

```
The MIT License (MIT)

Copyright (c) 2016

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
