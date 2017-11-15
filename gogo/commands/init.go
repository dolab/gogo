package commands

import (
	"log"

	"github.com/dolab/logger"
)

var (
	stderr *logger.Logger

	envTemplate = `#!/usr/bin/env bash

export APPROOT=$(pwd)

# adjust GOPATH
case ":$GOPATH:" in
    *":$APPROOT:"*) :;;
    *) GOPATH=$APPROOT:$GOPATH;;
esac
export GOPATH


# adjust PATH
readopts="ra"
if [ -n "$ZSH_VERSION" ]; then
    readopts="rA";
fi
while IFS=':' read -$readopts ARR; do
    for i in "${ARR[@]}"; do
        case ":$PATH:" in
            *":$i/bin:"*) :;;
            *) PATH=$i/bin:$PATH
        esac
    done
done <<< "$GOPATH"
export PATH


# mock development && test envs
if [ ! -d "$APPROOT/src/{{.Namespace}}/{{.Application}}" ];
then
    mkdir -p "$APPROOT/src/{{.Namespace}}"
    ln -s "$APPROOT/gogo/" "$APPROOT/src/{{.Namespace}}/{{.Application}}"
fi
`

	makefileTemplate = `all: gobuild gotest

godev:
    cd gogo && go run main.go

gobuild: goclean goinstall

gorebuild: goclean goreinstall

goclean:
    go clean ./...

goinstall:
    go get -t -v {{.Namespace}}/{{.Application}}

goreinstall:
    go get -t -a -v {{.Namespace}}/{{.Application}}

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

	modelTemplate = `
package models

import (
    "errors"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
    {{.Name}} *_{{.Name}}

    {{.LowerCaseName}}Collection = "{{.LowerCaseName}}"
    {{.LowerCaseName}}Indexes = []mgo.Index{}
)

type _{{.Name}} struct{}

type {{.Name}}Model struct {
    ID bson.ObjectId    ` + "`" + `bson:"_id"` + "`" + ` 
    CreatedAt time.Time ` + "`" + `bson:"created_at"` + "`" + `
    UpdatedAt time.Time ` + "`" + `bson:"updated_at"` + "`" + `

    isNewRecord bool   ` + "`" + `bson:"-"` + "`" + `
}

func New{{.Name}}Model() *{{.Name}}Model {
    return &{{.Name}}Model{
        ID:          bson.NewObjectId(),
        isNewRecord: true,
    }
}

func ({{.LowerCaseName}} *{{.Name}}Model) IsNewRecord() bool {
    return {{.LowerCaseName}}.isNewRecord
}

func ({{.LowerCaseName}} *{{.Name}}Model) Save() (err error) {
    if !{{.LowerCaseName}}.ID.Valid() {
        err = errors.New("Invalid bson id")
        return
    }

    {{.Name}}.Query(func(c *mgo.Collection){
        if {{.LowerCaseName}}.IsNewRecord() {
            err = c.Insert({{.LowerCaseName}})
            if err == nil {
                {{.LowerCaseName}}.isNewRecord = false
            }
        } else {
            update := bson.M{}

            err = c.UpdateId({{.LowerCaseName}}.ID, bson.M{ 
               "$set":  update,
            })
        }
    })

    return
}

func (_ *_{{.Name}}) Find(id string) ({{.LowerCaseName}} *{{.Name}}Model, err error) {
    if !bson.IsObjectIdHex(id) {
        err = errors.New("Invalid bson id")
        return
    }

    query := bson.M{
        "_id": bson.ObjectIdHex(id),
    }

    {{.Name}}.Query(func(c *mgo.Collection){
        err = c.Find(query).One(&{{.LowerCaseName}})
    })
    
    return
}

func ({{.LowerCaseName}} *{{.Name}}Model) Remove() (err error) {
    if !{{.LowerCaseName}}.ID.Valid() {
        return ErrInvalidID
    }

    {{.Name}}.Query(func(c *mgo.Collection) {
        err = c.RemoveId({{.LowerCaseName}}.ID)
    })

    return
}

func (_ *_{{.Name}}) Query(query func(c *mgo.Collection)) {
    mongo.Query({{.LowerCaseName}}Collection, {{.LowerCaseName}}Indexes, query)
}
`

	modelTestTemplate = `
package models

import(
    "testing"
    "github.com/golib/assert"
)

func Test_{{.Name}}(t *testing.T) {
    assertion := assert.New(t)

    model := New{{.Name}}Model()
    assertion.True(model.IsNewRecord())

    err := model.Save()
    assertion.Nil(err)
    assertion.False(model.IsNewRecord())

    res, err := {{.Name}}.Find(model.ID.Hex())
    assertion.Nil(err)

    err = res.Remove()
    assertion.Nil(err)
}
`

	controllerTemplate = `
package controllers

import (
	"net/http"

	"github.com/dolab/gogo"
)

var (
    {{.Name}} *_{{.Name}}
)

func (_ *_{{.Name}}) ID() string {
    return "id"
}

type _{{.Name}} struct{}

func (_ *_{{.Name}}) Index(ctx *gogo.Context) {
    ctx.SetStatus(http.StatusNotImplemented)
    ctx.Return()
}

func (_ *_{{.Name}}) Create(ctx *gogo.Context) {
    ctx.SetStatus(http.StatusNotImplemented)
    ctx.Return()
}

func (_ *_{{.Name}}) Show(ctx *gogo.Context) {
    ctx.SetStatus(http.StatusNotImplemented)
    ctx.Return()
}

func (_ *_{{.Name}}) Update(ctx *gogo.Context) {
    ctx.SetStatus(http.StatusNotImplemented)
    ctx.Return()
}

func (_ *_{{.Name}}) Destroy(ctx *gogo.Context) {
    ctx.SetStatus(http.StatusNotImplemented)
    ctx.Return()
}	
`

	controllerTestTemplate = `
package controllers

import(
    "testing"
)

func Test_Create_{{.Name}}(t *testing.T){

}

func Test_Index_{{.Name}}(t *testing.T) {
    
}

func Test_Show_{{.Name}}(t *testing.T) {

}

func Test_Update_{{.Name}}(t *testing.T) {
    
}

func Test_Destroy_{{.Name}}(t *testing.T) {

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
