package gogo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dolab/gogo/internal/listeners"
	"github.com/dolab/gogo/pkgs/hooks"
	"golang.org/x/net/http2"
)

var (
	network2addr = regexp.MustCompile(`(?i)^(http|https|tcp|tcp4|tcp6|unix|unixpacket):/{1,3}?(.+)?`)
)

// AppServer defines a server component of gogo
type AppServer struct {
	*AppGroup
	*hooks.ServerHooks

	config       Configer
	logger       Logger
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

	// init ServerHooks
	server.ServerHooks = hooks.NewServerHooks()

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
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.localAddr
}

// RegisterMiddlewares tries to register all middlewares defined
func (s *AppServer) RegisterMiddlewares(svc Servicer) {
	if registry, ok := svc.(RequestReceivedMiddlewarer); ok {
		for _, m := range registry.RequestReceived() {
			s.RegisterRequestReceived(m)
		}
	}

	if registry, ok := svc.(RequestRoutedMiddlewarer); ok {
		for _, m := range registry.RequestRouted() {
			s.RegisterRequestReceived(m)
		}
	}

	if registry, ok := svc.(ResponseReadyMiddlewarer); ok {
		for _, m := range registry.ResponseReady() {
			s.RegisterRequestReceived(m)
		}
	}

	if registry, ok := svc.(ResponseAlwaysMiddlewarer); ok {
		for _, m := range registry.ResponseAlways() {
			s.RegisterRequestReceived(m)
		}
	}
}

// RegisterRequestReceived registers middleware at request received phase
func (s *AppServer) RegisterRequestReceived(m Middlewarer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	name := m.Name()
	if s.RequestReceived.Has(name) {
		return fmt.Errorf("Middleware conflict, %q has registered for request received phase", name)
	}

	applier, err := m.Register(s.config.Middlewares())
	if err != nil {
		return err
	}

	s.RequestReceived.PushBackNamed(hooks.NamedHook{
		Name:     name,
		Apply:    applier,
		Priority: m.Priority(),
	})

	return nil
}

// RegisterRequestRouted registers middleware at request routed phase
func (s *AppServer) RegisterRequestRouted(m Middlewarer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	name := m.Name()
	if s.RequestRouted.Has(name) {
		return fmt.Errorf("Middleware conflict, %q has registered for request routed phase", name)
	}

	applier, err := m.Register(s.config.Middlewares())
	if err != nil {
		return err
	}

	s.RequestRouted.PushBackNamed(hooks.NamedHook{
		Name:     name,
		Apply:    applier,
		Priority: m.Priority(),
	})

	return nil
}

// RegisterResponseReady registers middleware at response ready phase
func (s *AppServer) RegisterResponseReady(m Middlewarer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	name := m.Name()
	if s.ResponseReady.Has(name) {
		return fmt.Errorf("Middleware conflict, %q has registered for response ready phase", name)
	}

	applier, err := m.Register(s.config.Middlewares())
	if err != nil {
		return err
	}

	s.ResponseReady.PushBackNamed(hooks.NamedHook{
		Name:     name,
		Apply:    applier,
		Priority: m.Priority(),
	})

	return nil
}

// RegisterResponseAlways registers middleware at response always phase
func (s *AppServer) RegisterResponseAlways(m Middlewarer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	name := m.Name()
	if s.ResponseAlways.Has(name) {
		return fmt.Errorf("Middleware conflict, %q has registered for response always phase", name)
	}

	applier, err := m.Register(s.config.Middlewares())
	if err != nil {
		return err
	}

	s.ResponseAlways.PushBackNamed(hooks.NamedHook{
		Name:     name,
		Apply:    applier,
		Priority: m.Priority(),
	})

	return nil
}

// NewService register all resources of service with middlewares
func (s *AppServer) NewService(svc Servicer) *AppServer {
	// try to register all middlewares
	s.RegisterMiddlewares(svc)

	svc.Init(s.config, s.AppGroup)

	// register filters
	svc.Filters()

	// register resources
	svc.Resources()

	return s
}

// NewResources register all resources of service without middlewares
func (s *AppServer) NewResources(svc Servicer) *AppServer {
	// try to register all middlewares
	s.RegisterMiddlewares(svc)

	svc.Init(s.config, s.AppGroup)

	// register resources
	svc.Resources()

	return s
}

// Run starts the http server with AppGroup as http.Handler
//
// NOTE: Run apply throughput and concurrency to http.Server.
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

	// register healthz
	if config.Server.Healthz {
		s.registerHealthz()
	}

	// throughput of rate limit
	if config.Server.Throttle > 0 {
		s.RequestReceived.PushFrontNamed(
			hooks.NewServerThrottleHook(config.Server.Throttle),
		)
	}

	// concurrency of bucket token
	if config.Server.Demotion > 0 {
		s.RequestReceived.PushBackNamed(
			hooks.NewServerDemotionHook(config.Server.Demotion, config.Server.RTimeout),
		)
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

	// register server
	log := s.loggerNew("GOGO")

	listener := listeners.New(config.Server.HTTP2)
	s.RequestReceived.PushFrontNamed(listener.RequestReceivedHook())

	conn, err := listener.Listen(network, addr)
	if err != nil {
		log.Fatalf("listeners.Listen(%s, %s): %v", network, addr, err)
	}
	log.Infof("Listened on %s://%s", network, addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           s.AppGroup,
		ReadHeaderTimeout: time.Duration(rtimeout) * time.Second,
		ReadTimeout:       time.Duration(rtimeout) * time.Second,
		WriteTimeout:      time.Duration(wtimeout) * time.Second,
		MaxHeaderBytes:    maxHeaderBytes,
	}
	server.RegisterOnShutdown(listener.Shutdown)

	// register locals
	s.mux.Lock()
	s.localAddr = addr
	s.localServ = server
	s.mux.Unlock()

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
	s.localSig = make(chan os.Signal, 1)
	signal.Notify(s.localSig, os.Interrupt)

	go s.Run()

	<-s.localSig
	close(s.localSig)

	log := s.loggerNew("GOGO")
	log.Info("Shutting down server ....")

	// NOTE: 500*time.Millinsecond is copied from net/http
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
