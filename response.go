package gogo

import (
	"log"
	"net/http"
)

const (
	noneHeaderFlushed = -1
)

// Response extends http.ResponseWriter with extra metadata
type Response struct {
	http.ResponseWriter

	status  int
	size    int
	filters []ResponseFilter
}

// NewResponse returns a Responser with w passed.
// NOTE: It sets default response status code to http.StatusOK
func NewResponse(w http.ResponseWriter) Responser {
	response := &Response{
		ResponseWriter: w,

		status: http.StatusOK,
		size:   noneHeaderFlushed,
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
	return r.size != noneHeaderFlushed
}

// WriteHeader sets response status code only by overwrites underline
func (r *Response) WriteHeader(code int) {
	if code > 0 {
		r.status = code

		if r.HeaderFlushed() {
			log.Println("[WARN] ", ErrHeaderFlushed.Error())
		}
	}
}

// FlushHeader writes response headers with status and reset size, it also invoke before filters
func (r *Response) FlushHeader() {
	if r.HeaderFlushed() {
		return
	}

	r.size = 0
	r.ResponseWriter.WriteHeader(r.status)
}

// Write writes data to client, and returns size of data written or error if exits.
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

// Flush tryes flush to client if possible
func (r *Response) Flush() {
	flush, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flush.Flush()
	}
}

// Reset resets the current *Response with new http.ResponseWriter
func (r *Response) Reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = http.StatusOK
	r.size = noneHeaderFlushed
	r.filters = nil
}
