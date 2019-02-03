package gogo

import (
	"crypto"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dolab/gogo/internal/params"
	"github.com/dolab/gogo/internal/render"
	"github.com/dolab/gogo/pkgs/hooks"
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return NewContext()
		},
	}

	// contextNew returns a new context for the request
	contextNew = func(w http.ResponseWriter, r *http.Request, ps *params.Params, pkg, ctrl, action string) *Context {
		ctx := contextPool.Get().(*Context)

		ctx.Response.Hijack(w)
		ctx.Request = r
		ctx.Params = ps
		ctx.Logger = NewContextLogger(r)
		ctx.pkg = pkg
		ctx.ctrl = ctrl
		ctx.action = action

		return ctx
	}

	// contextReuse puts the context back to pool for later usage
	contextReuse = func(ctx *Context) {
		contextPool.Put(ctx)
	}
)

// Context defines context of a request for gogo.
type Context struct {
	Response Responser
	Request  *http.Request
	Params   *params.Params
	Logger   Logger

	mux            sync.RWMutex
	settings       map[string]interface{}
	frozenSettings map[string]interface{}
	responseReady  *hooks.HookList
	cursor         int8

	pkg      string
	ctrl     string
	action   string
	filters  []FilterFunc
	issuedAt time.Time
}

// NewContext returns a *Context without initialization
func NewContext() *Context {
	return &Context{
		Response: NewResponse(nil),
		cursor:   -1,
	}
}

// Package returns package path of routed request.
func (c *Context) Package() string {
	return c.pkg
}

// Controller returns controller name of routed request.
func (c *Context) Controller() string {
	return c.ctrl
}

// Action returns action name of routed request.
func (c *Context) Action() string {
	return c.action
}

// Set binds a new value with key for the context
func (c *Context) Set(key string, value interface{}) {
	c.mux.Lock()

	if c.settings == nil {
		c.settings = make(map[string]interface{})
	}

	c.settings[key] = value

	c.mux.Unlock()
}

// Get returns a value of the key
func (c *Context) Get(key string) (v interface{}, ok bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.settings == nil {
		return
	}

	v, ok = c.settings[key]
	return
}

// MustGet returns a value of key or panic when key doesn't exist
func (c *Context) MustGet(key string) interface{} {
	v, ok := c.Get(key)
	if !ok {
		c.Logger.Panicf("Key %s doesn't exist", key)
	}

	return v
}

// SetFinal binds a value with key for the context and freezes it
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
func (c *Context) GetFinal(key string) (v interface{}, ok bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.frozenSettings == nil {
		return
	}

	v, ok = c.frozenSettings[key]
	return
}

// MustSetFinal likes SetFinal, but it panics if key is duplicated.
func (c *Context) MustSetFinal(key string, value interface{}) {
	err := c.SetFinal(key, value)
	if err != nil {
		c.Logger.Panicf("Freeze key %s duplicated!", key)
	}
}

// MustGetFinal returns a frozen value of key or panic when it doesn't exist
func (c *Context) MustGetFinal(key string) interface{} {
	v, ok := c.GetFinal(key)
	if !ok {
		c.Logger.Panicf("Freeze key %s doesn't exist", key)
	}

	return v
}

// RequestID returns request id of the Context
func (c *Context) RequestID() string {
	if c.Logger == nil {
		return ""
	}

	return c.Logger.RequestID()
}

// RequestURI returns request raw uri of http.Request
func (c *Context) RequestURI() string {
	if c.Request.RequestURI != "" {
		return c.Request.RequestURI
	}

	uri, _ := url.QueryUnescape(c.Request.URL.EscapedPath())
	return uri
}

// HasRawHeader returns true if request sets its header with specified key
func (c *Context) HasRawHeader(key string) bool {
	for yek := range c.Request.Header {
		if key == yek {
			return true
		}
	}

	return false
}

// RawHeader returns request header value of specified key
func (c *Context) RawHeader(key string) string {
	for yek, val := range c.Request.Header {
		if key == yek {
			return strings.Join(val, ",")
		}
	}

	return ""
}

// HasHeader returns true if request sets its header for canonicaled specified key
func (c *Context) HasHeader(key string) bool {
	_, ok := c.Request.Header[http.CanonicalHeaderKey(key)]

	return ok
}

// Header returns request header value of canonicaled specified key
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

// SetHeader sets response header with key/value pair
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// Redirect returns a HTTP redirect to the specific location.
func (c *Context) Redirect(location string) {
	// always abort
	c.Abort()

	// adjust status code, default to 302
	status := c.Response.Status()
	switch status {
	case http.StatusMovedPermanently, http.StatusSeeOther, http.StatusTemporaryRedirect:
		// skip

	default:
		status = http.StatusFound
	}

	c.SetHeader("Location", location)

	http.Redirect(c.Response, c.Request, location, status)
}

// Return returns response with auto detected Content-Type
func (c *Context) Return(body ...interface{}) error {
	var tmprender render.Render

	// auto detect response content-type from request header of accept
	if len(c.Response.Header().Get("Content-Type")) == 0 {
		accept := c.Request.Header.Get("Accept")
		for _, enc := range strings.Split(accept, ",") {
			switch {
			case strings.Contains(enc, render.ContentTypeJSON), strings.Contains(enc, render.ContentTypeJSONP):
				tmprender = render.NewJsonRender(c.Response)

			case strings.Contains(enc, render.ContentTypeXML):
				tmprender = render.NewXmlRender(c.Response)

			}
			if tmprender != nil {
				break
			}
		}
	}

	// third, use default render
	if tmprender == nil {
		tmprender = render.NewDefaultRender(c.Response)
	}

	if len(body) > 0 {
		return c.Render(tmprender, body[0])
	}

	return c.Render(tmprender, nil)
}

// HashedReturn returns response with ETag header calculated hash of response.Body dynamically
func (c *Context) HashedReturn(hasher crypto.Hash, body ...interface{}) error {
	if len(body) > 0 {
		return c.Render(render.NewHashRender(c.Response, hasher), body[0])
	}

	return c.Render(render.NewHashRender(c.Response, hasher), "")
}

// Text returns response with Content-Type: text/plain header
func (c *Context) Text(data interface{}) error {
	return c.Render(render.NewTextRender(c.Response), data)
}

// String is alias of Text
func (c *Context) String(data interface{}) error {
	return c.Render(render.NewTextRender(c.Response), data)
}

// JSON returns response with json codec and Content-Type: application/json header
func (c *Context) JSON(data interface{}) error {
	return c.Render(render.NewJsonRender(c.Response), data)
}

// JSONP returns response with json codec and Content-Type: application/javascript header
func (c *Context) JSONP(callback string, data interface{}) error {
	return c.Render(render.NewJsonpRender(c.Response, callback), data)
}

// XML returns response with xml codec and Content-Type: text/xml header
func (c *Context) XML(data interface{}) error {
	return c.Render(render.NewXmlRender(c.Response), data)
}

// Render responses client with data rendered by Render
func (c *Context) Render(rr render.Render, data interface{}) error {
	// always abort
	c.Abort()

	// currect response status code
	if coder, ok := data.(StatusCoder); ok {
		c.SetStatus(coder.StatusCode())
	}

	// currect response content-type header
	if contentType := rr.ContentType(); contentType != "" {
		c.SetHeader("Content-Type", contentType)
	}

	// flush header
	c.Response.FlushHeader()

	// invoke ResponseReady
	if !c.responseReady.Run(c.Response, c.Request) {
		return nil
	}

	// shortcut for nil data
	if data == nil {
		return nil
	}

	err := rr.Render(data)
	if err != nil {
		c.Logger.Errorf("%T.Render(?): %v", rr, err)
	}

	return err
}

// Next executes the remain filters in the chain.
//
// NOTE: It ONLY used in the filters!
func (c *Context) Next() {
	// is aborted?
	if c.cursor >= math.MaxInt8 {
		return
	}

	c.cursor++

	if c.cursor >= 0 && c.cursor < int8(len(c.filters)) {
		c.filters[c.cursor](c)
	}
}

// Abort forces to stop call chain.
func (c *Context) Abort() {
	c.cursor = math.MaxInt8
}

// run starting request chan with new envs.
func (c *Context) run(handler http.Handler, filters []FilterFunc, responseReady *hooks.HookList) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// reset internal
	c.settings = nil
	c.frozenSettings = nil
	c.filters = filters
	c.responseReady = responseReady
	c.issuedAt = time.Now()
	c.cursor = -1

	// start chains
	c.Next()

	// ghost for non render
	if c.cursor >= 0 && c.cursor < math.MaxInt8 {
		c.Abort()

		// invoke ResponseReady
		if !c.responseReady.Run(c.Response, c.Request) {
			return
		}

		// invoke http.Handler if defined
		if handler != nil {
			handler.ServeHTTP(c.Response, c.Request)
		} else {
			// response status code always
			c.Response.FlushHeader()
		}
	}
}
