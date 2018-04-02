package gogo

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"github.com/dolab/httpdispatch"
)

// AppParams defines params component of gogo
type AppParams struct {
	mux     sync.RWMutex
	request *http.Request
	params  httpdispatch.Params
	rawBody []byte
	rawErr  error
	readed  bool
}

// NewParams returns an *AppParams with *http.Request and context httpdispatch.Params
func NewParams(r *http.Request) *AppParams {
	return NewAppParams(r, httpdispatch.ContextParams(r))
}

// NewAppParams returns an *AppParams with *http.Request and httpdispatch.Params
func NewAppParams(r *http.Request, params httpdispatch.Params) *AppParams {
	return &AppParams{
		request: r,
		params:  params,
	}
}

// HasQuery returns whether named param is exist for URL query string.
func (p *AppParams) HasQuery(name string) bool {
	_, ok := p.request.URL.Query()[name]

	return ok
}

// HasForm returns whether named param is exist for POST/PUT request body.
func (p *AppParams) HasForm(name string) bool {
	p.request.ParseMultipartForm(DefaultMaxMultiformBytes)

	_, ok := p.request.PostForm[name]

	return ok
}

// RawBody returns request body and error if exist while reading.
func (p *AppParams) RawBody() ([]byte, error) {
	p.mux.RLock()
	if p.readed {
		p.mux.RUnlock()

		return p.rawBody, p.rawErr
	}
	p.mux.RUnlock()

	p.mux.Lock()
	defer p.mux.Unlock()

	p.rawBody, p.rawErr = ioutil.ReadAll(p.request.Body)

	// close the request.Body
	if p.rawErr == nil {
		p.request.Body.Close()
	}

	p.readed = true

	return p.rawBody, p.rawErr
}

// Get returns the first value for the named component of the request.
// NOTE: httpdispatch.Params takes precedence over URL query string values.
func (p *AppParams) Get(name string) string {
	value := p.params.ByName(name)

	// trye URL query string if value of route is empty
	if value == "" {
		value = p.request.URL.Query().Get(name)
	}

	return value
}

func (p *AppParams) GetInt(name string) (int, error) {
	return strconv.Atoi(p.Get(name))
}

func (p *AppParams) GetInt8(name string) (int8, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 8)
	return int8(result), err
}

func (p *AppParams) GetUint8(name string) (uint8, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 8)
	return uint8(result), err
}

func (p *AppParams) GetInt16(name string) (int16, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 16)
	return int16(result), err
}

func (p *AppParams) GetUint16(name string) (uint16, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 16)
	return uint16(result), err
}

func (p *AppParams) GetUint32(name string) (uint32, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 32)
	return uint32(result), err
}

func (p *AppParams) GetInt32(name string) (int32, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 32)
	return int32(result), err
}

func (p *AppParams) GetInt64(name string) (int64, error) {
	return strconv.ParseInt(p.Get(name), 10, 64)
}

func (p *AppParams) GetUint64(name string) (uint64, error) {
	return strconv.ParseUint(p.Get(name), 10, 64)
}

func (p *AppParams) GetFloat(name string) (float64, error) {
	return strconv.ParseFloat(p.Get(name), 64)
}

func (p *AppParams) GetBool(name string) (bool, error) {
	return strconv.ParseBool(p.Get(name))
}

// Post returns the named comonent of the request by calling http.Request.FormValue()
func (p *AppParams) Post(name string) string {
	return p.request.FormValue(name)
}

// File retrieves multipart uploaded file of HTTP POST request
func (p *AppParams) File(name string) (multipart.File, *multipart.FileHeader, error) {
	return p.request.FormFile(name)
}

// Json unmarshals request body with json codec
func (p *AppParams) Json(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// Xml unmarshals request body with xml codec
func (p *AppParams) Xml(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}

// Gob decode request body with gob codec
func (p *AppParams) Gob(v interface{}) error {
	data, err := p.RawBody()
	if err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}
