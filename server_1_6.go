// +build !go1.7

package gogo

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// ServeHTTP implements the http.Handler interface with throughput and concurrency support.
// NOTE: It servers client by dispatching request to AppRoute.
func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hijack request id
	requestID := r.Header.Get(s.requestID)
	if requestID == "" || len(requestID) > DefaultMaxHttpRequestIDLen {
		requestID = NewGID().Hex()

		// inject request header with new request id
		r.Header.Set(s.requestID, requestID)
	}

	logger := s.logger.New(requestID)
	defer s.logger.Reuse(logger)

	logger.Debugf(`processing %s "%s"`, r.Method, s.filterParameters(r.URL))

	// throughput by rate limit, timeout after time.Second/throttle
	if s.throttle != nil {
		ctx, done := context.WithTimeout(context.Background(), s.throttleTimeout)
		err := s.throttle.Wait(ctx)
		done()

		if err != nil {
			logger.Warnf("Throughput exceed: %v", err)

			w.Header().Set("Retry-After", s.throttleTimeout.String())
			http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
			return
		}
	}

	// concurrency by channel, timeout after request+response timeouts
	if s.slowdown != nil {
		ticker := time.NewTicker(s.slowdownTimeout)

		select {
		case <-s.slowdown:
			ticker.Stop()

			defer func() {
				s.slowdown <- true
			}()

		case <-ticker.C:
			ticker.Stop()

			logger.Warnf("Concurrency exceed: %v timeout", s.slowdownTimeout)

			w.Header().Set("Retry-After", s.slowdownTimeout.String())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
	}

	s.AppRoute.ServeHTTP(w, r)
}
