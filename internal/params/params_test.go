package params

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

func Test_ParamsHasQuery(t *testing.T) {
	it := assert.New(t)

	request, _ := http.NewRequest("GET", "/path/to/resource?test&key=url_value", nil)

	p := NewParams(request, httpdispatch.Params{})
	it.True(p.HasQuery("test"))
	it.True(p.HasQuery("key"))
	it.False(p.HasQuery("un-existed-key"))
}

func Test_ParamsHasForm(t *testing.T) {
	it := assert.New(t)

	params := url.Values{}
	params.Add("key", "name")

	request, _ := http.NewRequest("PUT", "/path/to/resource?test&key=url_value", bytes.NewBufferString(params.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	p := NewParams(request, httpdispatch.Params{})
	it.True(p.HasForm("key"))
	it.False(p.HasForm("un-existed-key"))
}

func Test_ParamsRawBody(t *testing.T) {
	it := assert.New(t)

	params := url.Values{}
	params.Add("key", "name")

	request, _ := http.NewRequest("POST", "/path/to/resource?test&key=url_value", bytes.NewBufferString(params.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	p := NewParams(request, httpdispatch.Params{})

	body, err := p.RawBody()
	if it.Nil(err) {
		it.Equal(params.Encode(), string(body))
	}

	// original FormValue should works also
	it.Equal(params.Get("key"), request.FormValue("key"))

	// safe to invoke more than one time
	body, err = p.RawBody()
	if it.Nil(err) {
		it.Equal(params.Encode(), string(body))
	}

	// original FormValue should works also
	it.Equal(params.Get("key"), request.FormValue("key"))
}

func Test_ParamsGet(t *testing.T) {
	it := assert.New(t)

	request, _ := http.NewRequest("GET", "/path/to/resource?key=url_value&test=url_true", nil)
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"key", "route_value"},
		httpdispatch.Param{"test", ""},
	}

	p := NewParams(request, routeParams)
	it.Equal("route_value", p.Get("key"))
	it.Equal("url_true", p.Get("test"))
	it.Empty(p.Get("un-existed-key"))
}

func Test_ParamsForm(t *testing.T) {
	it := assert.New(t)

	params := url.Values{}
	params.Add("key", "post_value")

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}

	p := NewParams(request, routeParams)
	it.Equal("post_value", p.Form("key"))
	it.Equal("url_true", p.Form("test"))
}

func fakeForm() (filename, contentType string, buf *bytes.Buffer) {
	var (
		root, _ = os.Getwd()
	)

	filename = root + "/const.go"
	buf = bytes.NewBuffer(nil)

	// build multipart form
	w := multipart.NewWriter(buf)
	w.WriteField("key", "form_value")
	fd, _ := os.Open(filename)
	form, _ := w.CreateFormFile("file", filename)
	io.Copy(form, fd)
	w.Close()

	contentType = w.FormDataContentType()

	return
}

func Test_ParamsFile(t *testing.T) {
	it := assert.New(t)

	filename, contentType, buf := fakeForm()

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", buf)
	request.Header.Set("Content-Type", contentType)
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}

	p := NewParams(request, routeParams)
	it.Equal("url_value", p.Form("key"))
	it.Equal("url_true", p.Form("test"))

	// validate uploaded content
	f, fh, err := p.File("file")
	if it.Nil(err) {
		it.Equal(filename, fh.Filename)

		buf.Reset()
		io.Copy(buf, f)

		content, err := ioutil.ReadFile(filename)
		if it.Nil(err) {
			it.Equal(content, buf.Bytes())
		}
	}
}

func Test_ParamsJson(t *testing.T) {
	it := assert.New(t)

	str := `{"key":"json_value", "test":true}`

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(str))
	request.Header.Set("Content-Type", "application/json")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}

	p := NewParams(request, routeParams)

	var params struct {
		Key  string `json:"key"`
		Test bool   `json:"test"`
	}
	err := p.Json(&params)
	if it.Nil(err) {
		it.Equal("json_value", params.Key)
		it.True(params.Test)
	}
}

func Test_ParamsXml(t *testing.T) {
	it := assert.New(t)

	str := `<Params><Key>xml_value</Key><Test>true</Test></Params>`

	request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(str))
	request.Header.Set("Content-Type", "application/json")
	routeParams := httpdispatch.Params{
		httpdispatch.Param{"test", "route_true"},
	}

	p := NewParams(request, routeParams)

	var params struct {
		XMLName xml.Name `xml:"Params"`
		Key     string   `xml:"Key"`
		Test    bool     `xml:"Test"`
	}
	err := p.Xml(&params)
	if it.Nil(err) {
		it.Equal("xml_value", params.Key)
		it.True(params.Test)
	}
}

func Test_ParamsGob(t *testing.T) {
	it := assert.New(t)

	type data struct {
		Key  string `bson:"key"`
		Test bool   `bson:"test"`
	}

	var buf bytes.Buffer

	params := data{
		Key:  "gob_value",
		Test: true,
	}

	err := gob.NewEncoder(&buf).Encode(params)
	if it.Nil(err) {
		request, _ := http.NewRequest("POST", "/path/to/resource?key=url_value&test=url_true", strings.NewReader(buf.String()))
		request.Header.Set("Content-Type", "application/gob")

		p := NewParams(request, httpdispatch.Params{})

		var temp data

		err = p.Gob(&temp)
		if it.Nil(err) {
			it.Equal("gob_value", temp.Key)
			it.True(temp.Test)
		}
	}
}
