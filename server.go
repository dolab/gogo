package gogo

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golib/httprouter"
)

type AppServer struct {
	*AppRoute

	mode   RunMode
	config *AppConfig
	logger *AppLogger
	router *httprouter.Router
	pool   sync.Pool

	requestId    string
	filterParams []string
}

func NewAppServer(mode RunMode, config *AppConfig, logger *AppLogger) *AppServer {
	server := &AppServer{
		mode:      mode,
		config:    config,
		logger:    logger,
		requestId: DefaultHttpRequestId,
	}

	// init Route
	server.AppRoute = NewAppRoute("/", server)

	// init router
	server.router = httprouter.New()
	server.router.HandleMethodNotAllowed = false // strict for RESTful
	// server.router.NotFound = http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
	// 	server.logger.Warnf("format", ...)
	// })

	// overwrite
	server.pool.New = func() interface{} {
		return NewContext(server)
	}

	return server
}

// Mode returns run mode of the app server
func (s *AppServer) Mode() string {
	return string(s.mode)
}

// Config returns app config of the app server
func (s *AppServer) Config() *AppConfig {
	return s.config
}

// Run runs the http server
func (s *AppServer) Run() {
	var (
		config = s.config.Section()

		network        = "tcp"
		addr           = config.Server.Addr
		port           = config.Server.Port
		rtimeout       = DefaultHttpRequestTimeout
		wtimeout       = DefaultHttpResponseTimeout
		maxheaderbytes = 0

		localAddr string
	)

	// adjust app server request id
	if config.Server.RequestId != "" {
		s.requestId = config.Server.RequestId
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
			s.logger.Fatal("=> SSL is only supported for TCP sockets.")
		}

		s.logger.Fatal("=> Failed to listen:", server.ListenAndServeTLS(config.Server.SslCert, config.Server.SslKey))
	} else {
		listener, err := net.Listen(network, localAddr)
		if err != nil {
			s.logger.Fatal("=> Failed to listen:", err)
		}

		s.logger.Fatal("=> Failed to serve:", server.Serve(listener))
	}
}

// Use dispatch to Route.Use by overwrite
func (s *AppServer) Use(middlewares ...Middleware) {
	s.AppRoute.Use(middlewares...)
}

// Clean removes all registered middlewares, it useful in testing cases.
func (s *AppServer) Clean() {
	s.Handlers = []Middleware{}
}

// ServeHTTP implements the http.Handler interface
func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Debugf(`processing %s "%s"`, r.Method, s.filterParameters(r.URL))

	s.router.ServeHTTP(w, r)
}

// New returns a new context for the server
func (s *AppServer) new(w http.ResponseWriter, r *http.Request, params *AppParams, handlers []Middleware) *Context {
	// adjust request id
	requestId := r.Header.Get(s.requestId)
	if requestId == "" {
		requestId = NewObjectId().Hex()

		// inject request header with new request id
		r.Header.Set(s.requestId, requestId)
	}
	w.Header().Set(s.requestId, requestId)

	ctx := s.pool.Get().(*Context)
	ctx.Request = r
	ctx.Response = &ctx.writer
	ctx.Params = params
	ctx.Logger = s.logger.New(requestId)
	ctx.settings = nil
	ctx.frozenSettings = nil
	ctx.writer.reset(w)
	ctx.handlers = handlers
	ctx.index = -1

	return ctx
}

// Reuse puts the context back to pool for later usage
func (s *AppServer) reuse(ctx *Context) {
	s.pool.Put(ctx)
}
