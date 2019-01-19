package gogo

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dolab/gogo/internal/listeners"
	"github.com/dolab/gogo/pkgs/gid"
	"github.com/dolab/gogo/pkgs/hooks"
	"golang.org/x/net/http2"
)

var (
	network2addr = regexp.MustCompile(`(?i)^(http|https|tcp|tcp4|tcp6|unix|unixpacket):/{1,3}?(.+)?`)
)

// AppServer defines a server component of gogo
type AppServer struct {
	*AppGroup

	config       Configer
	logger       Logger
	hooks        hooks.ServerHooks
	requestID    string   // request id header name
	filterFields []string // filter out params when logging

	localSig  chan os.Signal
	localAddr string
	localServ *http.Server
}

// NewAppServer returns *AppServer inited with args
func NewAppServer(config Configer, logger Logger) *AppServer {
	server := &AppServer{
		config:    config,
		logger:    logger,
		requestID: DefaultRequestIDKey,
	}

	// init AppGroup for server
	server.AppGroup = NewAppGroup("/", server)

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

// Address returns app listen address
func (s *AppServer) Address() string {
	return s.localAddr
}

// ServeHTTP implements the http.Handler interface with throughput and concurrency support.
// NOTE: It serves client by forwarding request to AppGroup.
func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hijack request id if required
	var logID string
	if s.hasRequestID() {
		logID = r.Header.Get(s.requestID)
		if logID == "" || len(logID) > DefaultRequestIDMaxLen {
			logID = gid.New().Hex()

			// inject request header with new request id
			r.Header.Set(s.requestID, logID)
		}

		w.Header().Set(s.requestID, logID)
	}

	// hijack logger for request
	log := s.loggerNew(logID)
	defer s.loggerReuse(log)

	r = r.WithContext(context.WithValue(r.Context(), ctxLoggerKey, log))

	if !s.hooks.RequestReceived.Run(w, r) {
		return
	}

	s.AppGroup.Serve(w, r)
}

// Run starts the http server with httpdispatch.Router handler
func (s *AppServer) Run() {
	var (
		config         = s.config.Section()
		network        = "tcp"
		addr           = config.Server.Addr
		port           = config.Server.Port
		rtimeout       = DefaultRequestTimeout
		wtimeout       = DefaultResponseTimeout
		maxHeaderBytes = 0
	)

	// throughput of rate limit
	if config.Server.Throttle > 0 {
		s.hooks.RequestReceived.PushFrontNamed(hooks.NewServerThrottleHook(config.Server.Throttle))
	}

	// concurrency of bucket token
	if config.Server.Slowdown > 0 {
		s.hooks.RequestReceived.PushBackNamed(hooks.NewServerDemotionHook(config.Server.Slowdown, config.Server.RTimeout))
	}

	// adjust app server request id if specified
	if config.Server.RequestID != "" {
		s.requestID = config.Server.RequestID
	}

	// adjust app logger filter sensitive fields
	s.filterFields = config.Logger.FilterFields

	// If the port is zero, treat the address as a fully qualified local address.
	// This address must be prefixed with the network type followed by a colon,
	// e.g. unix:/tmp/gogo.socket or tcp6:::1 (equivalent to tcp6:0:0:0:0:0:0:0:1)
	matches := network2addr.FindStringSubmatch(addr)
	if len(matches) == 3 {
		switch strings.ToLower(matches[1]) {
		case "http", "https":
			// ignore
		default:
			network = matches[1]
		}

		addr = "/" + strings.TrimPrefix(matches[2], "/")
	}

	if port != 0 {
		addr += ":" + strconv.Itoa(port)
	}

	if config.Server.RTimeout > 0 {
		rtimeout = config.Server.RTimeout
	}
	if config.Server.WTimeout > 0 {
		wtimeout = config.Server.WTimeout
	}
	if config.Server.MaxHeaderBytes > 0 {
		maxHeaderBytes = config.Server.MaxHeaderBytes
	}

	listener := listeners.New(config.Server.HTTP2)
	s.hooks.RequestReceived.PushFrontNamed(listener.RequestReceivedHook())

	log := s.loggerNew("GOGO")

	conn, err := listener.Listen(network, addr)
	if err != nil {
		log.Fatalf("net.Listen(%s, %s): %v", network, addr, err)
	}
	log.Infof("Listened on %s:%s", network, addr)

	server := &http.Server{
		Addr:           s.localAddr,
		Handler:        s,
		ReadTimeout:    time.Duration(rtimeout) * time.Second,
		WriteTimeout:   time.Duration(wtimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
	server.RegisterOnShutdown(listener.Shutdown)

	// register locals
	s.localAddr = addr
	s.localServ = server

	if config.Server.Ssl {
		msg := "ServeTLS(%s:%s): %v"
		if config.Server.HTTP2 {
			err := http2.ConfigureServer(server, nil)
			if err != nil {
				log.Fatalf("http2.ConfigureServer(%T, nil): %v", server, err)
			}

			msg = "ServeHTTP2(%s:%s): %v"
		}

		if err := server.ServeTLS(conn, config.Server.SslCert, config.Server.SslKey); err != nil {
			if strings.Contains(err.Error(), "http: Server closed") {
				log.Info("Server shutdown")
			} else {
				log.Fatalf(msg, network, addr, err)
			}
		}
	} else {
		if err := server.Serve(conn); err != nil {
			if strings.Contains(err.Error(), "http: Server closed") {
				log.Info("Server shutdown")
			} else {
				log.Fatalf("Serve(%s:%s): %v", network, addr, err)
			}
		}
	}
}

// RunWithHandler runs the http server with given handler
// It's useful for embbedding third-party golang applications.
func (s *AppServer) RunWithHandler(handler Handler) {
	s.SetHandler(handler)

	s.Run()
}

// Serve runs a server with graceful shutdown feature
func (s *AppServer) Serve() {
	localSig := make(chan os.Signal, 1)
	signal.Notify(localSig, os.Interrupt)

	if len(localSig) == 0 {
		go s.Run()
	}

	<-localSig

	log := s.loggerNew("GOGO")
	log.Info("Shutting down server ....")

	ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	s.localServ.Shutdown(ctx)

	select {
	case <-ctx.Done():
		os.Exit(0)
	}
}

// Shutdown shuts down AppServer gracefully by emitting os.Interrupt
func (s *AppServer) Shutdown() {
	if s.localSig == nil {
		return
	}

	// use interrupt sig
	s.localSig <- os.Interrupt
}

func (s *AppServer) loggerNew(tag string) Logger {
	return s.logger.New(tag)
}

func (s *AppServer) loggerReuse(l Logger) {
	s.logger.Reuse(l)
}

func (s *AppServer) hasRequestID() bool {
	return len(s.requestID) > 0
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
