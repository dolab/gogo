package params

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/dolab/httpdispatch"
)

// Params defines params component of gogo
type Params struct {
	mux     sync.RWMutex
	request *http.Request
	params  httpdispatch.Params
	rawBody []byte
	rawErr  error
	readed  bool
}

// New returns an *Params with *http.Request and context httpdispatch.Params
func New(r *http.Request) *Params {
	return NewParams(r, httpdispatch.ContextParams(r))
}

// NewParams returns an *Params with *http.Request and httpdispatch.Params
func NewParams(r *http.Request, params httpdispatch.Params) *Params {
	return &Params{
		request: r,
		params:  params,
	}
}

// HasQuery returns whether named param is exist for URL query string.
func (p *Params) HasQuery(name string) bool {
	_, ok := p.request.URL.Query()[name]

	return ok
}

// HasForm returns whether named param is exist for POST/PUT request body.
func (p *Params) HasForm(name string) bool {
	p.request.ParseMultipartForm(DefaultMaxMultiformBytes)

	_, ok := p.request.PostForm[name]

	return ok
}

// RawBody returns request body and error if exist while reading.
func (p *Params) RawBody() ([]byte, error) {
	p.mux.RLock()
	if p.readed {
		p.mux.RUnlock()

		return p.rawBody, p.rawErr
	}
	p.mux.RUnlock()

	p.mux.Lock()
	defer p.mux.Unlock()

	p.rawBody, p.rawErr = ioutil.ReadAll(p.request.Body)

	// close and hijack the request.Body
	if p.rawErr == nil {
		p.request.Body.Close()

		p.request.Body = ioutil.NopCloser(bytes.NewReader(p.rawBody))
	}

	// mark as readed
	p.readed = true

	return p.rawBody, p.rawErr
}

// Get returns the first value for the named component of the request.
// NOTE: httpdispatch.Params takes precedence over URL query string values.
func (p *Params) Get(name string) string {
	value := p.params.ByName(name)

	// trye URL query string if value of route is empty
	if value == "" {
		value = p.request.URL.Query().Get(name)
	}

	return value
}

// Form returns the named comonent of the request by calling http.Request.FormValue()
func (p *Params) Form(name string) string {
	return p.request.FormValue(name)
}

// File retrieves multipart uploaded file of HTTP POST request
func (p *Params) File(name string) (multipart.File, *multipart.FileHeader, error) {
	return p.request.FormFile(name)
}

// Json unmarshals request body with json codec
func (p *Params) Json(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// Xml unmarshals request body with xml codec
func (p *Params) Xml(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}

// Gob decode request body with gob codec
func (p *Params) Gob(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}
