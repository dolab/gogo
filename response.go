package gogo

import (
	"log"
	"net/http"
)

const (
	noneHeaderFlushed = -1
)

type Response struct {
	http.ResponseWriter

	status  int
	size    int
	filters []ResponseFilter
}

func NewResponse(w http.ResponseWriter) Responser {
	response := &Response{
		ResponseWriter: w,

		status: http.StatusOK,
		size:   noneHeaderFlushed,
	}

	return response
}

// Before registers filters which invoked before response has written
func (r *Response) Before(filter func(w Responser)) {
	r.filters = append(r.filters, filter)
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int {
	return r.size
}

// HeaderFlushed returns whether response headers has written
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

	// filters
	for i := len(r.filters) - 1; i >= 0; i-- {
		r.filters[i](r)
	}

	r.size = 0
	r.ResponseWriter.WriteHeader(r.status)
}

func (r *Response) Write(data []byte) (size int, err error) {
	r.FlushHeader()

	size, err = r.ResponseWriter.Write(data)

	r.size += size
	return
}

func (r *Response) Flush() {
	flush, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flush.Flush()
	}
}

func (r *Response) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = http.StatusOK
	r.size = noneHeaderFlushed
	r.filters = nil
}
