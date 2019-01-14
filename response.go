package gogo

import (
	"log"
	"net/http"
)

const (
	nonHeaderFlushed = -1
)

// Response extends http.ResponseWriter with extra metadata.
type Response struct {
	http.ResponseWriter

	filters []ResponseFilter
	status  int
	size    int
}

// NewResponse returns a Responser with w given.
// NOTE: It sets response status code to http.StatusNotImplemented by default.
func NewResponse(w http.ResponseWriter) Responser {
	response := &Response{
		ResponseWriter: w,
		status:         http.StatusNotImplemented,
		size:           nonHeaderFlushed,
	}

	return response
}

// Before registers filters which invoked before response has written
func (r *Response) Before(filter ResponseFilter) {
	r.filters = append(r.filters, filter)
}

// Status returns current response status code
func (r *Response) Status() int {
	return r.status
}

// Size returns size of response body written
func (r *Response) Size() int {
	return r.size
}

// HeaderFlushed returns true if response headers has written
func (r *Response) HeaderFlushed() bool {
	return r.size != nonHeaderFlushed
}

// WriteHeader sets response status code by overwriting underline
func (r *Response) WriteHeader(code int) {
	if code > 0 {
		r.status = code

		if r.HeaderFlushed() {
			log.Println("[WARN] ", ErrHeaderFlushed.Error())
		}
	}
}

// FlushHeader writes response headers of status code, and resets size of response.
func (r *Response) FlushHeader() {
	if r.HeaderFlushed() {
		return
	}

	r.size = 0
	r.ResponseWriter.WriteHeader(r.status)
}

// Write writes data to client, and returns size of data written.
// It returns error when failed.
// It also invokes before filters if exist.
func (r *Response) Write(data []byte) (size int, err error) {
	// apply filters
	for i := len(r.filters) - 1; i >= 0; i-- {
		data = r.filters[i](r, data)
	}

	r.FlushHeader()

	size, err = r.ResponseWriter.Write(data)

	r.size += size
	return
}

// Flush tryes flushing to client if possible
func (r *Response) Flush() {
	flush, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flush.Flush()
	}
}

// Hijack resets the current *Response with new http.ResponseWriter
func (r *Response) Hijack(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = http.StatusNotImplemented
	r.size = nonHeaderFlushed
	r.filters = nil
}
