package gogo

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golib/httprouter"
	"golang.org/x/time/rate"
)

// AppServer defines server component of gogo
type AppServer struct {
	*AppRoute

	pool    sync.Pool
	handler Handler

	throttle        *rate.Limiter // time.Rate for rate limit, its about throughput
	throttleTimeout time.Duration // time.Duration for throughput timeout
	slowdown        chan bool     // chain for rate limit, its about concurrency
	slowdownTimeout time.Duration // time.Duration for concurrency timeout

	config       *AppConfig
	logger       Logger
	requestID    string   // request id header name
	filterParams []string // filter out params when logging
}

// NewAppServer returns *AppServer inited with args
func NewAppServer(config *AppConfig, logger Logger) *AppServer {
	server := &AppServer{
		config:    config,
		logger:    logger,
		requestID: DefaultHttpRequestID,
	}

	// init Context pool
	server.pool.New = func() interface{} {
		return NewContext(server)
	}

	// init Route
	server.AppRoute = NewAppRoute("/", server)

	// init default handler with httprouter.Router
	handler := httprouter.New()
	handler.RedirectTrailingSlash = false
	handler.HandleMethodNotAllowed = false // strict for RESTful
	handler.NotFound = http.HandlerFunc(server.AppRoute.notFoundHandle)
	handler.MethodNotAllowed = http.HandlerFunc(server.AppRoute.methodNotAllowed)

	server.handler = handler

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

// Run runs the http server with httprouter.Router handler
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
		s.throttleTimeout = time.Second / time.Duration(config.Server.Throttle)
		s.throttle = rate.NewLimiter(rate.Every(s.throttleTimeout), config.Server.Throttle*10/100)
	}

	// concurrency of rate limit
	if config.Server.Slowdown > 0 {
		s.slowdownTimeout = time.Duration(config.Server.RTimeout) * time.Second
		s.slowdown = make(chan bool, config.Server.Slowdown)
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
	s.handler = handler

	s.Run()
}

// Use applies middlewares to app route
// NOTE: It dispatch to Route.Use by overwrite
func (s *AppServer) Use(middlewares ...Middleware) {
	s.AppRoute.Use(middlewares...)
}

// Handlers returns registered middlewares of app route
func (s *AppServer) Handlers() []Middleware {
	return s.AppRoute.handlers
}

// Clean removes all registered middlewares, it's useful in testing cases.
func (s *AppServer) Clean() {
	s.AppRoute.handlers = []Middleware{}
}

// ServeHTTP implements the http.Handler interface with throughput.
func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Debugf(`processing %s "%s"`, r.Method, s.filterParameters(r.URL))

	// throughput by rate limit, timeout after time.Second/throttle
	if s.throttle != nil {
		ctx, done := context.WithTimeout(context.Background(), s.throttleTimeout)
		err := s.throttle.Wait(ctx)
		done()

		if err != nil {
			s.logger.Warnf("Exceed server throughput: %v", err)

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

			s.logger.Warnf("Exceed server concurrency: %v timeout", s.slowdownTimeout)

			w.Header().Set("Retry-After", s.slowdownTimeout.String())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
	}

	s.handler.ServeHTTP(w, r)
}

// new returns a new context for the request
func (s *AppServer) newContext(r *http.Request, params *AppParams) *Context {
	// hijack request id
	requestID := r.Header.Get(s.requestID)
	if requestID == "" {
		requestID = NewGID().Hex()

		// inject request header with new request id
		r.Header.Set(s.requestID, requestID)
	}

	ctx := s.pool.Get().(*Context)
	ctx.Request = r
	ctx.Params = params
	ctx.Logger = s.logger.New(requestID)

	return ctx
}

// reuse puts the context back to pool for later usage
func (s *AppServer) reuseContext(ctx *Context) {
	s.logger.Reuse(ctx.Logger)

	s.pool.Put(ctx)
}
