package hooks

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// A ServerHooks provides a collection of server hooks for various
// stages of handling requests.
type ServerHooks struct {
	// RequestReceived is invoked as soon as a request enters the
	// server at the earliest available moment.
	RequestReceived HookList

	// RequestRouted is invoked when a request has been routed to a
	// particular handler of the server.
	RequestRouted HookList

	// ResponseReady is invoked when a request has been handled, between response header
	// has been flushed and response data is ready to be sent to the client.
	ResponseReady HookList

	// ResponseAlways is invoked when all bytes of a response (including an error
	// response) have been written.
	ResponseAlways HookList
}

// NewServerDebugLogHook returns a func for server debugging with log.
func NewServerDebugLogHook() func(HookItem) bool {
	return func(item HookItem) bool {
		return true
	}
}

// NewServerThrottleHook creates NamedHook with max throughput / per second.
// NOTE: burst value is 20% of throttle
func NewServerThrottleHook(max int) NamedHook {
	burst := max * 20 / 100
	if burst < 1 {
		burst = 1
	}

	timeout := time.Second / time.Duration(max)
	limiter := rate.NewLimiter(rate.Every(timeout), burst)

	return NamedHook{
		Name: "__server@throttle",
		Apply: func(w http.ResponseWriter, r *http.Request) bool {
			ctx, done := context.WithTimeout(context.Background(), timeout)
			err := limiter.Wait(ctx)
			done()

			if err != nil {
				log.Println("Exceed throughput:", err)

				w.Header().Set("Retry-After", time.Now().Add(timeout*3).Format(time.RFC3339))
				http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
				return false
			}

			return true
		},
	}
}

// NewServerDemotionHook creates NamedHook with max concurrency within seconds.
// NOTE: timeout after seconds
func NewServerDemotionHook(max, seconds int) NamedHook {
	if max < 1 {
		max = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	bucket := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		bucket <- struct{}{}
	}

	timeout := time.Duration(seconds) * time.Second

	// impulse sender
	burst := max / seconds
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(burst))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for i := 0; i < burst; i++ {
					bucket <- struct{}{}
				}
			}
		}
	}()

	return NamedHook{
		Name: "__server@demotion",
		Apply: func(w http.ResponseWriter, r *http.Request) bool {
			ticker := time.NewTicker(timeout)
			defer ticker.Stop()

			select {
			case <-bucket:
				return true

			case <-ticker.C:
				log.Println("Exceed concurrency:", timeout, "timeout")

				ticker.Stop()

				w.Header().Set("Retry-After", time.Now().Add(timeout).Format(time.RFC3339))
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

				return false
			}
		},
	}
}
