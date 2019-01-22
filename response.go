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

	status int
	size   int
}

// NewResponse returns a Responser with w given.
//
// NOTE: It sets response status code to http.StatusOK by default.
func NewResponse(w http.ResponseWriter) Responser {
	response := &Response{
		ResponseWriter: w,
		status:         http.StatusOK,
		size:           nonHeaderFlushed,
	}

	return response
}

// WriteHeader sets response status code by overwriting underline
func (r *Response) WriteHeader(code int) {
	if code > 0 {
		r.status = code

		if r.HeaderFlushed() {
			log.Println("[WARN]", ErrHeaderFlushed.Error())
		}
	}
}

// Write send data to client, and returns size of data written. It
// returns error of underline failed.
func (r *Response) Write(data []byte) (size int, err error) {
	r.FlushHeader()

	// shortcut for nil
	if data == nil {
		return
	}

	size, err = r.ResponseWriter.Write(data)

	r.size += size

	return
}

// Flush tryes flushing to client if possible
func (r *Response) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	} else {
		log.Println("[WARN] http.ResponseWriter does not implement http.Flusher")
	}
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

// FlushHeader writes response headers of status code, and resets size of response.
func (r *Response) FlushHeader() {
	if r.HeaderFlushed() {
		return
	}

	r.size = 0
	r.ResponseWriter.WriteHeader(r.status)
}

// Hijack resets the current *Response with new http.ResponseWriter
func (r *Response) Hijack(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = http.StatusOK
	r.size = nonHeaderFlushed
}
