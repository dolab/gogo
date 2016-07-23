# gogo

[![Build Status](https://travis-ci.org/dolab/gogo.svg?branch=master&style=flat)](https://travis-ci.org/dolab/gogo)

RESTful api framework of golang.

It's heavily inspired from [neko](https://github.com/rocwong/neko) which created by [RocWong](https://github.com/rocwong).

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

- Create application from skeleton

> **DEPRECATED**!!! Please using scaffold way.

```bash
$ cp -r $GOPATH/src/github.com/dolab/gogo/skeleton myapp

# fix application import path
$ cd myapp
$ source fix.sh
$ source env.sh

# run development server
$ make godev
```

## Getting Started

- Normal

```go
package main

import (
    "github.com/dolab/gogo"
)

func main() {
    app := gogo.New("development", "/path/to/your/config")

    // GET /
    app.GET("/", func(ctx *gogo.Context) {
        ctx.Text("Hello, gogo!")
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
    v1 := app.Group("/v1", func(ctx *gogo.Context) {
        auth := ctx.Header("Authorization")
        if !strings.HasPrefix(auth, "Basic ") {
            ctx.Abort()

            ctx.WriteHeader(http.StatusForbidden)
            ctx.Return()
            return
        }

        b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
        if err != nil {
            ctx.Logger.Errorf("Base64 decode error: %v", err)
            ctx.Abort()

            ctx.WriteHeader(http.StatusForbidden)
            ctx.Return()
            return
        }

        tmp := strings.SplitN(string(b), ":", 2)
        if len(tmp) != 2 || tmp[0] != "gogo" || tmp[1] != "ogog" {
            ctx.Abort()

            ctx.WriteHeader(http.StatusForbidden)
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
You can iml a struct with Index, Show, Create, Update, Destroy, Explore methods
and using AppServer.Resource("resource", your_struct) to register all it's impl methods to router
note: when your resource has a inheritance relationship, there must not be two same id key.
you can iml ControllerID interface to specify a new id key
```
type GroupController struct{}

func (t *GroupController) Index(ctx *Context) {
    ctx.Text(ctx.Params.Get("id") + "all")
}

func (t *GroupController) Show(ctx *Context) {
    ctx.Text(ctx.Params.Get("id") + "show")
}

type UserController struct{}

func (t *UserController) Show(ctx *Context) {
    ctx.Text(ctx.Params.Get("group") + ":" + ctx.Params.Get("user") + "show")
}

// iml ControllerID to specify user a new id key
func (t *UserController) Id() string {
    return "user"
}

group := app.Resource("group", &GroupController{})
user := group.("user", &UserController{)

```

## TODOs

- [x] server config context
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
