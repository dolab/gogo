package gogo

import (
	"bytes"
	"encoding/gob"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dolab/httpdispatch"
	"github.com/golib/assert"
)

func Test_AppParamsHasQuery(t *testing.T) {
	request, _ := http.NewRequest("GET", "/path/to/resource?test&key=url_value", nil)
	assertion := assert.New(t)

	p := NewAppParams(request, httpdispatch.Params{})
	assertion.True(p.HasQuery("test"))
	assertion.True(p.HasQuery("key"))
	assertion.False(p.HasQuery("un-existed-key"))
}

func Test_AppParamsHasForm(t *testing.T) {
	params := url.Values{}
	params.Add("key", "name")
	request, _ := http.NewRequest("PUT", "/path/to/resource?test&key=url_value", bytes.NewBufferString(params.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assertion := assert.New(t)

	p := NewAppParams(request, httpdispatch.Params{})
	assertion.True(p.HasForm("key"))
	assertion.False(p.HasForm("un-existed-key"))
}

func Test_AppParamsRawBody(t *testing.T) {
	params := url.Values{}
	params.Add("key", "name")

	request, _ := http.NewRequest("PUT", "/path/to/resource?test&key=url_value", bytes.NewBufferString(params.Encode()))
	assertion := assert.New(t)

	p := NewAppParams(request, httpdispatch.Params{})

	body, err := p.RawBody()
	assertion.Nil(err)
	assertion.Equal(params.Encode(), string(body))

	// safe to invoke more than one time
	body, err = p.RawBody()
	assertion.Nil(err)
	assertion.Equal(params.Encode(), string(body))
}

func Test_AppParamsGet(t *testing.T) {
	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"key", "route_value"},
		httpdispatch.Param{"test", ""},
	}
	assertion := assert.New(t)

	p := NewAppParams(request, routeParams)
	assertion.Equal("route_value", p.Get("key"))
	assertion.Equal("url_true", p.Get("test"))
	assertion.Empty(p.Get("un-existed-key"))
}

func Test_AppParamsPost(t *testing.T) {
	params := url.Values{}
	params.Add("key", "post_value")

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}
	assertion := assert.New(t)

	p := NewAppParams(request, routeParams)
	assertion.Equal("post_value", p.Post("key"))
	assertion.Equal("url_true", p.Post("test"))
}

func Test_AppParamsFile(t *testing.T) {
	var (
		root, _  = os.Getwd()
		filename = root + "/param.go"

		buf bytes.Buffer
	)

	// build multipart form
	w := multipart.NewWriter(&buf)
	w.WriteField("key", "form_value")
	fd, _ := os.Open(filename)
	form, _ := w.CreateFormFile("file", filename)
	io.Copy(form, fd)
	w.Close()

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", &buf)
	request.Header.Set("Content-Type", w.FormDataContentType())
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}
	assertion := assert.New(t)

	p := NewAppParams(request, routeParams)
	assertion.Equal("url_value", p.Post("key"))
	assertion.Equal("url_true", p.Post("test"))

	// validate uploaded content
	f, fh, err := p.File("file")
	assertion.Nil(err)
	assertion.Equal(filename, fh.Filename)

	buf.Reset()
	io.Copy(&buf, f)
	content, _ := ioutil.ReadFile(filename)
	assertion.Equal(content, buf.Bytes())
}

func Test_AppParamsJson(t *testing.T) {
	str := `{
    "key": "json_value",
    "test": true
}`

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(str))
	request.Header.Set("Content-Type", "application/json")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}
	assertion := assert.New(t)

	p := NewAppParams(request, routeParams)

	var params struct {
		Key  string `json:"key"`
		Test bool   `json:"test"`
	}
	err := p.Json(&params)
	assertion.Nil(err)
	assertion.Equal("json_value", params.Key)
	assertion.True(params.Test)
}

func Test_AppParamsXml(t *testing.T) {
	str := `<Params>
    <Key>xml_value</Key>
    <Test>true</Test>
</Params>`

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(str))
	request.Header.Set("Content-Type", "application/json")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}
	assertion := assert.New(t)

	p := NewAppParams(request, routeParams)

	var params struct {
		XMLName xml.Name `xml:"Params"`
		Key     string   `xml:"Key"`
		Test    bool     `xml:"Test"`
	}
	err := p.Xml(&params)
	assertion.Nil(err)
	assertion.Equal("xml_value", params.Key)
	assertion.True(params.Test)
}

func Test_AppParamsGob(t *testing.T) {
	type data struct {
		Key  string `bson:"key"`
		Test bool   `bson:"test"`
	}

	var buf bytes.Buffer

	assertion := assert.New(t)
	params := data{
		Key:  "gob_value",
		Test: true,
	}

	err := gob.NewEncoder(&buf).Encode(params)
	assertion.Nil(err)

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(buf.String()))
	request.Header.Set("Content-Type", "application/gob")

	p := NewAppParams(request, httpdispatch.Params{})

	var temp data
	err = p.Gob(&temp)
	assertion.Nil(err)
	assertion.Equal("gob_value", temp.Key)
	assertion.True(temp.Test)
}
