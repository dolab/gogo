package templates

var (
	componentControllerTemplate = `package controllers

import (
	"net/http"

	"github.com/dolab/gogo"
)

// @resources {{.Name}}
var (
	{{.Name}} *_{{.Name}}
)

type _{{.Name}} struct{}

// // custom resource id name, default to {{.Name|lowercase}}
// func (*_{{.Name}}) ID() string {
// 	return "id"
// }

// @route GET /{{.Name|lowercase}}
func (*_{{.Name}}) Index(ctx *gogo.Context) {
	ctx.SetStatus(http.StatusNotImplemented)
	ctx.Return()
}

// @route POST /{{.Name|lowercase}}
func (*_{{.Name}}) Create(ctx *gogo.Context) {
	ctx.SetStatus(http.StatusNotImplemented)
	ctx.Return()
}

// @route GET /{{.Name|lowercase}}/:{{.Name|lowercase}}
func (*_{{.Name}}) Show(ctx *gogo.Context) {
	// retrieve resource name of path params
	id := ctx.Params.Get("{{.Name|lowercase}}")

	ctx.SetStatus(http.StatusNotImplemented)
	ctx.Json(map[string]interface{}{
		"id": id,
		"tags": []string{
			id,
		},
	})
}

// @route PUT /{{.Name|lowercase}}/:{{.Name|lowercase}}
func (*_{{.Name}}) Update(ctx *gogo.Context) {
	// retrieve resource name of path params
	id := ctx.Params.Get("{{.Name|lowercase}}")

	ctx.SetStatus(http.StatusNotImplemented)
	ctx.Return(id)
}

// @route DELETE /{{.Name|lowercase}}/:{{.Name|lowercase}}
func (*_{{.Name}}) Destroy(ctx *gogo.Context) {
	// retrieve resource name of path params
	id := ctx.Params.Get("{{.Name|lowercase}}")

	ctx.SetStatus(http.StatusNotImplemented)
	ctx.Return(id)
}
`
	componentControllerTestTemplate = `package controllers

import (
	"net/http"
	"testing"
)

func Test_{{.Name}}_Index(t *testing.T) {
	request := gogotesting.New(t)
	request.Get("/{{.Name|lowercase}}")

	request.AssertStatus(http.StatusNotImplemented)
	request.AssertEmpty()
}

func Test_{{.Name}}_Create(t *testing.T) {
	request := gogotesting.New(t)
	request.PostJSON("/{{.Name|lowercase}}", nil)

	request.AssertStatus(http.StatusNotImplemented)
	request.AssertEmpty()
}

func Test_{{.Name}}_Show(t *testing.T) {
	id := "{{.Name|lowercase}}"

	request := gogotesting.New(t)
	request.Get("/{{.Name|lowercase}}/" + id)

	request.AssertStatus(http.StatusNotImplemented)
	request.AssertContainsJSON("id", id)
	request.AssertContainsJSON("tags.0", id)
}

func Test_{{.Name}}_Update(t *testing.T) {
	id := "{{.Name|lowercase}}"

	request := gogotesting.New(t)
	request.PutJSON("/{{.Name|lowercase}}/"+id, nil)

	request.AssertStatus(http.StatusNotImplemented)
	request.AssertContains(id)
}

func Test_{{.Name}}_Destroy(t *testing.T) {
	id := "{{.Name|lowercase}}"

	request := gogotesting.New(t)
	request.DeleteJSON("/{{.Name|lowercase}}/"+id, nil)

	request.AssertStatus(http.StatusNotImplemented)
	request.AssertContains(id)
}
`
	componentFilterTemplate = `package middlewares

import (
	"github.com/dolab/gogo"
)

func {{.Name}}() gogo.FilterFunc {
	return func(ctx *gogo.Context) {
		// TODO: implements custom logic
		ctx.AddHeader("x-gogo-filter", "Hello, Filter!")

		ctx.Next()
	}
}
`
	componentFilterTestTemplate = `package middlewares

import (
	"net/http"
	"testing"

	"github.com/dolab/gogo"
)

func Test_{{.Name}}(t *testing.T) {
	// register temp resource for testing
	app := gogoapp.NewGroup("", {{.Name}}())
	app.GET("/filters/{{.Name|lowercase}}", func(ctx *gogo.Context) {
		ctx.SetStatus(http.StatusNotImplemented)
	})

	request := gogotesting.New(t)
	request.Get("/filters/{{.Name|lowercase}}", nil)
	request.AssertStatus(http.StatusNotImplemented)
	request.AssertHeader("x-gogo-filter", "Hello, Filter!")
}
`
	componentModelTemplate = `package models

import (
	"errors"
)

var (
	{{.Name}} *_{{.Name}}
)

type {{.Name}}Model struct {
	// TODO: fill with table fields
}

func New{{.Name}}Model() *{{.Name}}Model {
	return &{{.Name}}Model{}
}

// helpers
type _{{.Name}} struct{}

func (*_{{.Name}}) Find(id string) (m *{{.Name}}Model, err error) {
	err = errors.New("Not Found")

	return
}
`
	componentModelTestTemplate = `package models

import (
	"testing"

	"github.com/golib/assert"
)

func Test_{{.Name}}Model(t *testing.T) {
	it := assert.New(t)

	m := New{{.Name}}Model()
	it.NotNil(m)
}

func Test_{{.Name}}_Find(t *testing.T) {
	it := assert.New(t)

	id := "???"

	m, err := {{.Name}}.Find(id)
	it.EqualError(err, "Not Found")
	it.Nil(m)
}
`
)
