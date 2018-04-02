// +build !go1.7

package gogo

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// NewContextHandle returns new *ContextHandle with handler info parsed
func NewContextHandle(server *AppServer, handler http.HandlerFunc, filters []Middleware) *ContextHandle {
	var rval reflect.Value

	if handler == nil {
		rval = reflect.ValueOf(filters[len(filters)-1])
	} else {
		rval = reflect.ValueOf(handler)
	}

	// formated in "/path/to/main.(*_Controller).Action-fm"
	name := runtime.FuncForPC(rval.Pointer()).Name()

	tmpvars := strings.Split(name, "/")
	if len(tmpvars) > 1 {
		name = tmpvars[len(tmpvars)-1]
	}

	tmpvars = strings.Split(name, ".")
	switch len(tmpvars) {
	case 3:
		// variable func in !go1.6 formatted in /path/to/gogo.glob.func?
		if tmpvars[1] == "glob" && tmpvars[2][:4] == "func" {
			tmpvars = []string{tmpvars[0], tmpvars[0], "<http.HandlerFunc>"}
		} else {
			// adjust controller name
			tmpvars[1] = strings.TrimLeft(tmpvars[1], "(")
			tmpvars[1] = strings.TrimRight(tmpvars[1], ")")

			// adjust action name
			tmpvars[2] = strings.SplitN(tmpvars[2], "-", 2)[0]
		}

	case 2:
		// package func
		tmpvars = []string{tmpvars[0], tmpvars[0], tmpvars[1]}

	default:
		tmpvars = []string{tmpvars[0], tmpvars[0], "<http.HandlerFunc>"}
	}

	return &ContextHandle{
		pkg:     tmpvars[0],
		ctrl:    tmpvars[1],
		action:  tmpvars[2],
		server:  server,
		handler: handler,
		filters: filters,
	}
}
