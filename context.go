package gogo

import (
	"math"
	"net/http"
	"sync"
)

const (
	abortIndex = math.MaxInt8 / 2
)

type Context struct {
	Response Responser
	Request  *http.Request
	Params   *AppParams

	Server *AppServer
	Config *AppConfig
	Logger *AppLogger

	mux            sync.RWMutex
	settings       map[string]interface{}
	frozenSettings map[string]interface{}

	writer   Response
	handlers []Middleware
	index    int8
}

func NewContext(server *AppServer) *Context {
	return &Context{
		Server: server,
		index:  -1,
	}
}

// Set associates a new value with key for the context
func (c *Context) Set(key string, value interface{}) {
	c.mux.Lock()

	if c.settings == nil {
		c.settings = make(map[string]interface{})
	}

	c.settings[key] = value

	c.mux.Unlock()
}

// Get returns a value of the key
func (c *Context) Get(key string) (interface{}, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.settings == nil {
		return nil, false
	}

	value, ok := c.settings[key]
	return value, ok
}

// MustGet returns a value of key or panic when key doesn't exist
func (c *Context) MustGet(key string) interface{} {
	value, ok := c.Get(key)
	if !ok {
		c.Logger.Panicf("Key %s doesn't exist", key)
	}

	return value
}

// SetFinal associates a value with key for the context and freezes following update
func (c *Context) SetFinal(key string, value interface{}) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.frozenSettings == nil {
		c.frozenSettings = make(map[string]interface{})
	}

	if _, ok := c.frozenSettings[key]; ok {
		return ErrSettingsKey
	}

	c.frozenSettings[key] = value

	return nil
}

// GetFinal returns a frozen value of the key
func (c *Context) GetFinal(key string) (interface{}, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.frozenSettings == nil {
		return nil, false
	}

	value, ok := c.frozenSettings[key]
	return value, ok
}

// MustSetFinal associates a value with key for the context and freezes following update,
// it panics if key is duplicated.
func (c *Context) MustSetFinal(key string, value interface{}) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.frozenSettings == nil {
		c.frozenSettings = make(map[string]interface{})
	}

	if _, ok := c.frozenSettings[key]; ok {
		c.Logger.Panicf("Freeze key %s duplicated!", key)
	}

	c.frozenSettings[key] = value
}

// MustGetFinal returns a frozen value of key or panic when key doesn't exist
func (c *Context) MustGetFinal(key string) interface{} {
	value, ok := c.GetFinal(key)
	if !ok {
		c.Logger.Panicf("Freeze key %s doesn't exist", key)
	}

	return value
}

// HasHeader returns true if request sets its header for specified key
func (c *Context) HasHeader(key string) bool {
	key = http.CanonicalHeaderKey(key)

	_, ok := c.Request.Header[key]

	return ok
}

// Header returns request header value of specified key
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetStatus sets response status code
func (c *Context) SetStatus(code int) {
	c.Response.WriteHeader(code)
}

// AddHeader adds response header with key/value pair
func (c *Context) AddHeader(key, value string) {
	c.Response.Header().Add(key, value)
}

// SetHeaders sets response header with key/value pair
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// Next executes the remain handlers in the chain.
// NOTE: It ONLY used in the middlewares!
func (c *Context) Next() {
	c.index++

	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)

		c.index++
	}
}

// Abort forces to stop call chain.
func (c *Context) Abort() {
	c.index = abortIndex
}

// Redirect returns a HTTP redirect to the specific location.
func (c *Context) Redirect(location string) {
	c.SetHeader("Location", location)

	// adjust status code, default to 302
	status := c.Response.Status()
	switch status {
	case http.StatusMovedPermanently, http.StatusSeeOther, http.StatusTemporaryRedirect:
		// skip

	default:
		status = http.StatusFound
	}

	http.Redirect(c.Response, c.Request, location, status)
}

// Return returns response following default Content-Type header
func (c *Context) Return(body ...interface{}) {
	if len(body) > 0 {
		c.Render(NewDefaultRender(c.Response), body[0])
	} else {
		c.Render(NewDefaultRender(c.Response), "")
	}
}

// Text returns response with Content-Type: text/plain header
func (c *Context) Text(data interface{}) {
	c.Render(NewTextRender(c.Response), data)
}

// Json returns response with json codec and Content-Type: application/json header
func (c *Context) Json(data interface{}) {
	c.Render(NewJsonRender(c.Response), data)
}

// JsonP returns response with json codec and Content-Type: application/javascript header
func (c *Context) JsonP(callback string, data interface{}) {
	c.Render(NewJsonpRender(c.Response, callback), data)
}

// Xml returns response with xml codec and Content-Type: text/xml header
func (c *Context) Xml(data interface{}) {
	c.Render(NewXmlRender(c.Response), data)
}

func (c *Context) Render(w Render, data interface{}) {
	err := w.Render(data)
	if err == nil {
		return
	}

	c.Logger.Errorf("%T render data with %v", w, err)
	c.Abort()

	c.Response.WriteHeader(http.StatusInternalServerError)
}
