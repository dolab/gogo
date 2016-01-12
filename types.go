package gogo

import "net/http"

// Middleware represents request filters and resource handler
type Middleware func(ctx *Context)

// Render represents HTTP response render
type Render interface {
	ContentType() string
	Render(v interface{}) error
}

type StatusCoder interface {
	StatusCode() int
}

// Responser represents HTTP response interface
type Responser interface {
	http.ResponseWriter
	http.Flusher

	Before(filter func(w Responser)) // register before filter
	Size() int                       // return the size of response body
	Status() int                     // response status code
	HeaderFlushed() bool             // whether response header has been sent?
	FlushHeader()                    // send response header
}

type ResponseFilter func(Responser)
