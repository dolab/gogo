package debugger

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/dolab/gogo/pkgs/middleware"
	yaml "gopkg.in/yaml.v2"
)

// A Debugger implements middleware.Interface
type Debugger struct {
	name  string
	phase middleware.Phase

	mux    sync.RWMutex
	config *Config
}

// New creates *Debugger with name
func New(phase middleware.Phase) *Debugger {
	if !phase.IsValid() {
		panic(middleware.ErrInvalidPhase)
	}

	return &Debugger{
		name:  "debugger",
		phase: phase,
	}
}

// Name returns name of debugger
func (d *Debugger) Name() string {
	return d.name
}

// Config returns settings template of debugger
func (d *Debugger) Config() []byte {
	b, _ := yaml.Marshal(Config{})

	return b
}

// Priority returns sort order of debugger
func (d *Debugger) Priority() int {
	return d.config.Priority
}

// Register unmarshals config of debbuger and return
func (d *Debugger) Register(unmarshaler middleware.Configer) (callee middleware.Interceptor, err error) {
	d.mux.Lock()
	defer d.mux.Unlock()

	if unmarshaler == nil {
		err = errors.New("invalid unmarshaler")
		return
	}

	err = unmarshaler.Unmarshal(d.Name(), &d.config)
	if err != nil {
		return
	}

	if !d.config.Debugable() {
		err = errors.New("no debugger")
		return
	}

	callee = d.interceptor

	return
}

// Reload tries to update settings of debugger at fly
func (d *Debugger) Reload(unmarshaler middleware.Configer) (err error) {
	d.mux.Lock()
	defer d.mux.Unlock()

	return unmarshaler.Unmarshal(d.Name(), &d.config)
}

// Shutdown will do nothing
func (d *Debugger) Shutdown() (err error) {
	return
}

var (
	logRequest = `DEBUG: Request %s/%s
---[ REQUEST DETAILS ]-------------------------------
%s
-----------------------------------------------------`

	logRequestError = `DEBUG ERROR: Request %s/%s
---[ REQUEST DUMP ERROR ]-----------------------------
%s
------------------------------------------------------`
)

func (d *Debugger) interceptor(w http.ResponseWriter, r *http.Request) bool {
	switch d.phase {
	case middleware.RequestReceived:
		if d.config.DebugRequest {
			data, err := httputil.DumpRequest(r, d.config.DebugRequestBody)
			if err != nil {
				log.Println(fmt.Sprintf(logRequestError, r.Method, r.RequestURI, err))
			} else {
				log.Println(fmt.Sprintf(logRequest, r.Method, r.RequestURI, data))
			}
		}

	case middleware.ResponseReady:
		// TODO: implements debugger for response?
	}

	// debugger always returns true
	return true
}
