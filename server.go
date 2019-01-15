package gogo

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dolab/gogo/pkgs/gid"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"golang.org/x/time/rate"
)

// AppServer defines server component of gogo
type AppServer struct {
	*AppRoute

	config       Configer
	logger       Logger
	requestID    string   // request id header name
	filterFields []string // filter out params when logging

	context         sync.Pool     // cache pool of *Context for performance
	throttle        *rate.Limiter // time.Rate for rate limit, its about throughput
	throttleTimeout time.Duration // time.Duration for throughput timeout
	slowdown        chan bool     // chain for rate limit, its about concurrency
	slowdownTimeout time.Duration // time.Duration for concurrency timeout
}

// NewAppServer returns *AppServer inited with args
func NewAppServer(config Configer, logger Logger) *AppServer {
	server := &AppServer{
		config:    config,
		logger:    logger,
		requestID: DefaultHttpRequestID,
	}

	// overwrite pool.New for pool of *Context
	server.context.New = func() interface{} {
		return NewContext()
	}

	// init AppRoute for server
	server.AppRoute = NewAppRoute("/", server)

	return server
}

// Mode returns run mode of the app server
func (s *AppServer) Mode() string {
	return s.config.RunMode().String()
}

// Config returns app config of the app server
func (s *AppServer) Config() Configer {
	return s.config
}

// ServeHTTP implements the http.Handler interface with throughput and concurrency support.
// NOTE: It servers client by dispatching request to AppRoute.
func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hijack request id if defined
	var requestID string
	if s.hasRequestID() {
		requestID = r.Header.Get(s.requestID)
		if requestID == "" || len(requestID) > DefaultMaxHttpRequestIDLen {
			requestID = gid.New().Hex()

			// inject request header with new request id
			r.Header.Set(s.requestID, requestID)
		}
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

// Run starts the http server with httpdispatch.Router handler
func (s *AppServer) Run() {
	var (
		config         = s.config.Section()
		network        = "tcp"
		addr           = config.Server.Addr
		port           = config.Server.Port
		rtimeout       = DefaultHttpRequestTimeout
		wtimeout       = DefaultHttpResponseTimeout
		maxheaderbytes = 0

		localAddr string
	)

	// throughput of rate limit
	// NOTE: burst value is 10% of throttle
	if config.Server.Throttle > 0 {
		s.newThrottle(config.Server.Throttle)
	}

	// concurrency of rate limit
	// NOTE: timeout after request timeout
	if config.Server.Slowdown > 0 {
		s.newSlowdown(config.Server.Slowdown, config.Server.RTimeout)
	}

	// adjust app server request id if specified
	if config.Server.RequestID != "" {
		s.requestID = config.Server.RequestID
	}

	// adjust app logger filter sensitive fields
	s.filterFields = config.Logger.FilterFields

	// If the port is zero, treat the address as a fully qualified local address.
	// This address must be prefixed with the network type followed by a colon,
	// e.g. unix:/tmp/app.socket or tcp6:::1 (equivalent to tcp6:0:0:0:0:0:0:0:1)
	if port == 0 {
		pieces := strings.SplitN(addr, ":", 2)

		network = pieces[0]
		localAddr = pieces[1]
	} else {
		localAddr = addr + ":" + strconv.Itoa(port)
	}

	if config.Server.RTimeout > 0 {
		rtimeout = config.Server.RTimeout
	}
	if config.Server.WTimeout > 0 {
		wtimeout = config.Server.WTimeout
	}
	if config.Server.MaxHeaderBytes > 0 {
		maxheaderbytes = config.Server.MaxHeaderBytes
	}

	server := &http.Server{
		Addr:           localAddr,
		Handler:        s,
		ReadTimeout:    time.Duration(rtimeout) * time.Second,
		WriteTimeout:   time.Duration(wtimeout) * time.Second,
		MaxHeaderBytes: maxheaderbytes,
	}

	s.logger.Infof("Listening on %s", localAddr)
	listener, err := net.Listen(network, localAddr)
	if err != nil {
		s.logger.Fatal("[GOGO]=> Failed to listen:", err)
	}

	s.startServer(server, listener, config.Server)
}

// RunWithHandler runs the http server with given handler
// It's useful for embbedding third-party golang applications.
func (s *AppServer) RunWithHandler(handler Handler) {
	s.SetHandler(handler)

	s.Run()
}

func (s *AppServer) filterParameters(lru *url.URL) string {
	ss := lru.Path

	query := lru.Query()
	if len(query) > 0 {
		for _, key := range s.filterFields {
			if _, ok := query[key]; ok {
				query.Set(key, "[FILTERED]")
			}
		}

		ss += "?" + query.Encode()
	}

	return ss
}

func (s *AppServer) startServer(server *http.Server, listener net.Listener, config *ServerConfig) {
	if config.Ssl {
		msg := "[GOGO]=> Failed to serve(TLS):"
		if config.HTTP2 {
			err := http2.ConfigureServer(server, nil)
			if err != nil {
				s.logger.Fatalf("[GOGO]=> http2.ConfigureServer(%T, nil): %v", server, err)
			}

			msg = "[GOGO]=> Failed to serve(HTTP2):"
		}

		s.logger.Fatal(msg, server.ServeTLS(listener, config.SslCert, config.SslKey))
	} else {
		s.logger.Fatal("[GOGO]=> Failed to serve:", server.Serve(listener))
	}
}

// new returns a new context for the request
func (s *AppServer) newContext(r *http.Request, controller, action string, params *AppParams) *Context {
	ctx := s.context.Get().(*Context)
	ctx.Request = r
	ctx.Params = params
	ctx.Logger = s.logger.New(r.Header.Get(s.requestID))
	ctx.controller = controller
	ctx.action = action

	return ctx
}

// reuse puts the context back to pool for later usage
func (s *AppServer) reuseContext(ctx *Context) {
	s.logger.Reuse(ctx.Logger)

	s.context.Put(ctx)
}

func (s *AppServer) hasRequestID() bool {
	return len(s.requestID) > 0
}

func (s *AppServer) newThrottle(n int) {
	burst := n * 10 / 100
	if burst < 1 {
		burst = 1
	}

	s.throttleTimeout = time.Second / time.Duration(n)
	s.throttle = rate.NewLimiter(rate.Every(s.throttleTimeout), burst)
}

func (s *AppServer) newSlowdown(n, timeout int) {
	s.slowdownTimeout = time.Duration(timeout) * time.Second
	s.slowdown = make(chan bool, n)
	for i := 0; i < n; i++ {
		s.slowdown <- true
	}
}
