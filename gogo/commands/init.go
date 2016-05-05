package commands

import (
	"log"

	"github.com/dolab/logger"
)

var (
	stderr *logger.Logger

	envTemplate = `#!/usr/bin/env bash

# adjust GOPATH
case ":$GOPATH:" in
    *":$(pwd):"*) :;;
    *) GOPATH=$(pwd):$GOPATH;;
esac
export GOPATH


# adjust PATH
readopts="-ra"
if [ -n "$ZSH_VERSION" ]; then
    readopts="-rA";
fi
while IFS=':' read -ra ARR; do
    for i in "${ARR[@]}"; do
        case ":$PATH:" in
            *":$i/bin:"*) :;;
            *) PATH=$i/bin:$PATH
        esac
    done
done <<< "$GOPATH"
export PATH


# mock development && test envs
if [ ! -d "$(pwd)/src/{{.Namespace}}/{{.Application}}" ];
then
    mkdir -p "$(pwd)/src/{{.Namespace}}"
    ln -s "$(pwd)/gogo/" "$(pwd)/src/{{.Namespace}}/{{.Application}}"
fi
`

	makefileTemplate = `all: gobuild gotest

godev:
    cd gogo && go run main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
    rm -rf bin
    rm -rf pkg

goinstall:
    cd gogo && go get -t -v ./...

goreinstall:
    cd gogo && go get -t -a -v ./...

gotest:
    go test {{.Namespace}}/{{.Application}}/app/controllers
    go test {{.Namespace}}/{{.Application}}/app/middlewares
    # go test {{.Namespace}}/{{.Application}}/app/models

gopackage:
    mkdir -p bin && go build -a -o bin/{{.Application}} src/{{.Namespace}}/{{.Application}}/main.go

travis: gobuild gotest
`

	gitIgnoreTemplate = `# Compiled Object files, Static and Dynamic libs (Shared Objects)
*.o
*.a
*.so

# Folders
_obj
_test
bin
pkg
src

# Architecture specific extensions/prefixes
*.[568vq]
[568vq].out

*.cgo1.go
*.cgo2.c
_cgo_defun.c
_cgo_gotypes.go
_cgo_export.*

_testmain.go

*.exe
*.test
*.prof

# development & test config files
*.development.json
*.test.json
`

	applicationTemplate = []string{`package controllers

import (
    "github.com/dolab/gogo"

    "{{.Namespace}}/{{.Application}}/app/middlewares"
)

type Application struct {
    *gogo.AppServer
}

func New(runMode, srcPath string) *Application {
    appServer := gogo.New(runMode, srcPath)

    err := NewAppConfig(appServer.Config())
    if err != nil {
        panic(err.Error())
    }

    return &Application{appServer}
}

// Middlerwares implements gogo.Middlewarer
// NOTE: DO NOT change the method name, its required by gogo!
func (app *Application) Middlewares() {
    // apply your middlewares

    // panic recovery
    app.Use(middlewares.Recovery())
}

// Resources implements gogo.Resourcer
// NOTE: DO NOT change the method name, its required by gogo!
func (app *Application) Resources() {
    // register your resources
    // app.GET("/", handler)

    app.GET("/@getting_start/hello", GettingStart.Hello)
}

// Run runs application after registering middelwares and resources
func (app *Application) Run() {
    // register middlewares
    app.Middlewares()

    // register resources
    app.Resources()

    // run server
    app.AppServer.Run()
}
`, `package controllers

import (
    "net/http/httptest"
    "os"
    "path"
    "testing"

    "github.com/dolab/httptesting"
)

var (
    testServer *httptest.Server
    testClient *httptesting.Client
)

func TestMain(m *testing.M) {
    var (
        runMode = "test"
        srcPath = path.Clean("../../")
    )

    app := New(runMode, srcPath)
    app.Resources()

    testServer = httptest.NewServer(app)
    testClient = httptesting.New(testServer.URL, false)

    code := m.Run()

    testServer.Close()

    os.Exit(code)
}
`}

	configTemplate = []string{`package controllers

import (
    "github.com/dolab/gogo"
)

var (
    Config *AppConfig
)

// Application configuration specs
type AppConfig struct {
    Domain       string              ` + "`" + `json:"domain"` + "`" + `
    GettingStart *GettingStartConfig ` + "`" + `json:"getting_start"` + "`" + `
}

// NewAppConfig apply application config from *gogo.AppConfig
func NewAppConfig(config *gogo.AppConfig) error {
    return config.UnmarshalJSON(&Config)
}

// Sample application config for illustration
type GettingStartConfig struct {
    Greeting string ` + "`" + `json:"greeting"` + "`" + `
}
`, `package controllers

import (
    "testing"

    "github.com/golib/assert"
)

func Test_AppConfig(t *testing.T) {
    assertion := assert.New(t)

    assertion.NotEmpty(Config.Domain)
    assertion.NotNil(Config.GettingStart)
}
`}

	gettingStartTemplate = []string{`package controllers

import (
    "github.com/dolab/gogo"
)

var (
    GettingStart *_GettingStart
)

type _GettingStart struct{}

// @route GET /@getting_start/hello
func (_ *_GettingStart) Hello(ctx *gogo.Context) {
    ctx.Logger.Warnf("Visiting domain is: %s", Config.Domain)

    ctx.Text(Config.GettingStart.Greeting)
}
`, `package controllers

import (
    "testing"
)

func Test_ExampleHello(t *testing.T) {
    testClient.Get(t, "/@getting_start/hello")

    testClient.AssertOK()
    testClient.AssertContains(Config.GettingStart.Greeting)
}
`}

	middlewareTemplate = []string{`package middlewares

import (
    "runtime"
    "strings"

    "github.com/dolab/gogo"
)

func Recovery() gogo.Middleware {
    return func(ctx *gogo.Context) {
        defer func() {
            if panicErr := recover(); panicErr != nil {
                // where does panic occur? try max 20 depths
                pcs := make([]uintptr, 20)
                max := runtime.Callers(2, pcs)
                for i := 0; i < max; i++ {
                    pcfunc := runtime.FuncForPC(pcs[i])
                    if strings.HasPrefix(pcfunc.Name(), "runtime.") {
                        continue
                    }

                    pcfile, pcline := pcfunc.FileLine(pcs[i])

                    tmp := strings.SplitN(pcfile, "/src/", 2)
                    if len(tmp) == 2 {
                        pcfile = "src/" + tmp[1]
                    }
                    ctx.Logger.Errorf("(%s:%d: %v)", pcfile, pcline, panicErr)

                    break
                }

                ctx.Abort()
            }
        }()

        ctx.Next()
    }
}
`, `package middlewares

import (
    "testing"

    "github.com/dolab/gogo"
)

func Test_Recovery(t *testing.T) {
    testApp.Use(Recovery())
    defer testApp.Clean()

    // register temp resource for testing
    testApp.GET("/middlewares/recovery", func(ctx *gogo.Context) {
        panic("Recover testing")
    })

    testClient.Get(t, "/middlewares/recovery", nil)
    testClient.AssertOK()
}
`, `package middlewares

import (
    "net/http/httptest"
    "os"
    "path"
    "testing"

    "github.com/dolab/gogo"
    "github.com/dolab/httptesting"
)

var (
    testApp    *gogo.AppServer
    testServer *httptest.Server
    testClient *httptesting.Client
)

func TestMain(m *testing.M) {
    var (
        runMode = "test"
        srcPath = path.Clean("../../")
    )

    testApp = gogo.New(runMode, srcPath)
    testServer = httptest.NewServer(testApp)
    testClient = httptesting.New(testServer.URL, false)

    code := m.Run()

    testServer.Close()

    os.Exit(code)
}
`}

	jsonTemplate = `{
    "name": "{{.Application}}",
    "mode": "test",
    "sections": {
        "development": {
            "server": {
                "addr": "localhost",
                "port": 9090,
                "ssl": false,
                "ssl_cert": "/path/to/ssl/cert",
                "ssl_key": "/path/to/ssl/key",
                "request_timeout": 30,
                "response_timeout": 30,
                "request_id": "X-Request-Id"
            },
            "logger": {
                "output": "stdout",
                "level": "debug",
                "filter_params": ["password", "password_confirmation"]
            },
            "domain": "https://example.com",
            "getting_start": {
                "greeting": "Hello, gogo!"
            }
        },

        "test": {
            "server": {
                "addr": "localhost",
                "port": 9090,
                "ssl": false,
                "ssl_cert": "/path/to/ssl/cert",
                "ssl_key": "/path/to/ssl/key",
                "request_timeout": 30,
                "response_timeout": 30,
                "request_id": "X-Request-Id"
            },
            "logger": {
                "output": "stdout",
                "level": "info",
                "filter_params": ["password", "password_confirmation"]
            },
            "domain": "https://example.com",
            "getting_start": {
                "greeting": "Hello, gogo!"
            }
        },

        "production": {
            "server": {
                "addr": "localhost",
                "port": 9090,
                "ssl": true,
                "ssl_cert": "/path/to/ssl/cert",
                "ssl_key": "/path/to/ssl/key",
                "request_timeout": 30,
                "response_timeout": 30,
                "request_id": "X-Request-Id"
            },
            "logger": {
                "output": "stdout",
                "level": "warn",
                "filter_params": ["password", "password_confirmation"]
            }
        }
    }
}
`

	mainTemplate = `package main

import (
    "flag"
    "os"
    "path"

    "github.com/dolab/gogo"

    "{{.Namespace}}/{{.Application}}/app/controllers"
)

var (
    runMode string // app run mode, available values are [development|test|production], default to development
    srcPath string // app source path, e.g. /home/deploy/websites/helloapp
)

func main() {
    flag.StringVar(&runMode, "runMode", "development", "{{.Application}} -runMode=[development|test|production]")
    flag.StringVar(&srcPath, "srcPath", "", "{{.Application}} -srcPath=/path/to/source")
    flag.Parse()

    // verify run mode
    if mode := gogo.RunMode(runMode); !mode.IsValid() {
        flag.PrintDefaults()
        return
    }

    // adjust src path
    if srcPath == "" {
        var err error

        srcPath, err = os.Getwd()
        if err != nil {
            panic(err)
        }
    } else {
        srcPath = path.Clean(srcPath)
    }

    controllers.New(runMode, srcPath).Run()
}
`
)

type templateData struct {
	Namespace   string
	Application string
}

func init() {
	var err error

	// setup logger
	stderr, err = logger.New("stderr")
	if err != nil {
		panic(err.Error())
	}

	stderr.SetLevelByName("info")
	stderr.SetFlag(log.Lshortfile)
}
