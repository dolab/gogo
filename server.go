package gogo

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// AppServer defines server component of gogo
type AppServer struct {
	*AppRoute

	config       *AppConfig
	logger       Logger
	requestID    string   // request id header name
	filterParams []string // filter out params when logging

	context         sync.Pool     // cache of *Context for performance
	throttle        *rate.Limiter // time.Rate for rate limit, its about throughput
	throttleTimeout time.Duration // time.Duration for throughput timeout
	slowdown        chan bool     // chain for rate limit, its about concurrency
	slowdownTimeout time.Duration // time.Duration for concurrency timeout
}

// NewAppServer returns *AppServer inited with args
func NewAppServer(config *AppConfig, logger Logger) *AppServer {
	server := &AppServer{
		config:    config,
		logger:    logger,
		requestID: DefaultHttpRequestID,
	}

	// overwrite pool.New for Context
	server.context.New = func() interface{} {
		return NewContext()
	}

	// init AppRoute for server
	server.AppRoute = NewAppRoute("/", server)

	return server
}

// Mode returns run mode of the app server
func (s *AppServer) Mode() string {
	return s.config.Mode.String()
}

// Config returns app config of the app server
func (s *AppServer) Config() *AppConfig {
	return s.config
}

// Run runs the http server with httpdispatch.Router handler
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

	// adjust app server request id
	if config.Server.RequestID != "" {
		s.requestID = config.Server.RequestID
	}

	// adjust app logger filter parameters
	s.filterParams = config.Logger.FilterParams

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
	if config.Server.Ssl {
		if network != "tcp" {
			// This limitation is just to reduce complexity, since it is standard
			// to terminate SSL upstream when using unix domain sockets.
			s.logger.Fatal("[GOGO]=> SSL is only supported for TCP sockets.")
		}

		s.logger.Fatal("[GOGO]=> Failed to listen:", server.ListenAndServeTLS(config.Server.SslCert, config.Server.SslKey))
	} else {
		listener, err := net.Listen(network, localAddr)
		if err != nil {
			s.logger.Fatal("[GOGO]=> Failed to listen:", err)
		}

		s.logger.Fatal("[GOGO]=> Failed to serve:", server.Serve(listener))
	}
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
		for _, key := range s.filterParams {
			if _, ok := query[key]; ok {
				query.Set(key, "[FILTERED]")
			}
		}

		ss += "?" + query.Encode()
	}

	return ss
}

// new returns a new context for the request
func (s *AppServer) newContext(r *http.Request, controller, action string, params *AppParams) *Context {
	ctx := s.context.Get().(*Context)
	ctx.Request = r
	ctx.controller = controller
	ctx.action = action
	ctx.Params = params
	ctx.Logger = s.logger.New(r.Header.Get(s.requestID))

	return ctx
}

// reuse puts the context back to pool for later usage
func (s *AppServer) reuseContext(ctx *Context) {
	s.logger.Reuse(ctx.Logger)

	s.context.Put(ctx)
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
