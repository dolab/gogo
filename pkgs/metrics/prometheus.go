package metrics

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/prometheus/client_golang/prometheus"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/peer"
// 	"google.golang.org/grpc/stats"
// )

// const userAgentLabel = "user_agent"

// // ServiceInfoProvider is simple wrapper around GetServiceInfo method.
// // This interface is implemented by grpc Server.
// type ServiceInfoProvider interface {
// 	// GetServiceInfo returns a map from service names to ServiceInfo.
// 	// Service names include the package names, in the form of <package>.<service>.
// 	GetServiceInfo() map[string]grpc.ServiceInfo
// }

// // RegisterInterceptor preallocates possible dimensions of every metric.
// // If peer tracking is enabled, nothing will happen.
// // If you register interceptor very frequently (for example during tests) it can allocate huge amount of memory.
// func RegisterInterceptor(s ServiceInfoProvider, i *Interceptor) (err error) {
// 	if i.trackPeers {
// 		return nil
// 	}

// 	infos := s.GetServiceInfo()
// 	for sn, info := range infos {
// 		for _, m := range info.Methods {
// 			t := handlerType(m.IsClientStream, m.IsServerStream)

// 			for c := uint32(0); c <= 15; c++ {
// 				requestLabels := prometheus.Labels{
// 					"service":      sn,
// 					"handler":      m.Name,
// 					"code":         codes.Code(c).String(),
// 					"type":         t,
// 					userAgentLabel: userAgent(context.TODO()),
// 				}
// 				messageLabels := prometheus.Labels{
// 					"service":      sn,
// 					"handler":      m.Name,
// 					userAgentLabel: userAgent(context.TODO()),
// 				}

// 				// server
// 				if _, err = i.monitoring.server.errors.GetMetricWith(requestLabels); err != nil {
// 					return fmt.Errorf("server errors metric initialization failure: %s", err.Error())
// 				}
// 				if _, err = i.monitoring.server.requestsTotal.GetMetricWith(requestLabels); err != nil {
// 					return fmt.Errorf("server requests total metric initialization failure: %s", err.Error())
// 				}
// 				if _, err = i.monitoring.server.requestDuration.GetMetricWith(requestLabels); err != nil {
// 					return fmt.Errorf("server requests duration metric initialization failure: %s", err.Error())
// 				}
// 				if m.IsClientStream {
// 					if _, err = i.monitoring.server.messagesReceived.GetMetricWith(messageLabels); err != nil {
// 						return fmt.Errorf("server messages received metric initialization failure: %s", err.Error())
// 					}
// 				}
// 				if m.IsServerStream {
// 					if _, err = i.monitoring.server.messagesSend.GetMetricWith(messageLabels); err != nil {
// 						return fmt.Errorf("server messages send metric initialization failure: %s", err.Error())
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// // Interceptor ...
// type Interceptor struct {
// 	monitoring *monitoring
// 	trackPeers bool
// }

// // InterceptorOpts ...
// type InterceptorOpts struct {
// 	// TrackPeers allow to turn on peer tracking.
// 	// For more info about peers please visit https://godoc.org/google.golang.org/grpc/peer.
// 	// peer is not bounded dimension so it can cause performance loss.
// 	// If its turn on Interceptor will not init metrics on startup.
// 	TrackPeers bool

// 	// ConstLabels will be passed to each collector.
// 	// Thanks to that it is possible to register multiple Interceptors.
// 	// They just have to have different constant labels.
// 	ConstLabels prometheus.Labels
// }

// // NewInterceptor implements both prometheus Collector interface and methods required by grpc Interceptor.
// func NewInterceptor(opts InterceptorOpts) *Interceptor {
// 	return &Interceptor{
// 		monitoring: initMonitoring(opts.TrackPeers, opts.ConstLabels),
// 		trackPeers: opts.TrackPeers,
// 	}
// }

// // Dialer ...
// func (i *Interceptor) Dialer(f func(string, time.Duration) (net.Conn, error)) func(string, time.Duration) (net.Conn, error) {
// 	return func(addr string, timeout time.Duration) (net.Conn, error) {
// 		i.monitoring.dialer.WithLabelValues(addr).Inc()
// 		return f(addr, timeout)
// 	}
// }

// // UnaryClient ...
// func (i *Interceptor) UnaryClient() grpc.UnaryClientInterceptor {
// 	monitor := i.monitoring.client

// 	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
// 		start := time.Now()

// 		err := invoker(ctx, method, req, reply, cc, opts...)
// 		code := grpc.Code(err)
// 		service, method := split(method)
// 		labels := prometheus.Labels{
// 			"service": service,
// 			"handler": method,
// 			"code":    code.String(),
// 			"type":    "unary",
// 		}
// 		if err != nil && code != codes.OK {
// 			monitor.errors.With(labels).Add(1)
// 		}

// 		monitor.requestDuration.With(labels).Observe(time.Since(start).Seconds())
// 		monitor.requestsTotal.With(labels).Add(1)

// 		return err
// 	}
// }

// // StreamClient ...
// func (i *Interceptor) StreamClient() grpc.StreamClientInterceptor {
// 	monitor := i.monitoring.client

// 	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
// 		start := time.Now()

// 		client, err := streamer(ctx, desc, cc, method, opts...)
// 		code := grpc.Code(err)
// 		service, method := split(method)
// 		labels := prometheus.Labels{
// 			"service": service,
// 			"handler": method,
// 			"code":    code.String(),
// 			"type":    handlerType(desc.ClientStreams, desc.ServerStreams),
// 		}
// 		if err != nil && code != codes.OK {
// 			monitor.errors.With(labels).Add(1)
// 		}

// 		monitor.requestDuration.With(labels).Observe(time.Since(start).Seconds())
// 		monitor.requestsTotal.With(labels).Add(1)

// 		return &monitoredClientStream{ClientStream: client, monitor: monitor, labels: prometheus.Labels{
// 			"service": service,
// 			"handler": method,
// 		}}, err
// 	}
// }

// // UnaryServer ...
// func (i *Interceptor) UnaryServer() grpc.UnaryServerInterceptor {
// 	monitor := i.monitoring.server

// 	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 		start := time.Now()

// 		res, err := handler(ctx, req)
// 		code := grpc.Code(err)
// 		service, method := split(info.FullMethod)

// 		labels := prometheus.Labels{
// 			"service":      service,
// 			"handler":      method,
// 			"code":         code.String(),
// 			"type":         "unary",
// 			userAgentLabel: userAgent(ctx),
// 		}
// 		if i.trackPeers {
// 			labels["peer"] = peerValue(ctx)
// 		}
// 		if err != nil && code != codes.OK {
// 			monitor.errors.With(labels).Add(1)
// 		}

// 		monitor.requestDuration.With(labels).Observe(time.Since(start).Seconds())
// 		monitor.requestsTotal.With(labels).Add(1)

// 		return res, err
// 	}
// }

// // StreamServer ...
// func (i *Interceptor) StreamServer() grpc.StreamServerInterceptor {
// 	monitor := i.monitoring.server

// 	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 		start := time.Now()
// 		service, method := split(info.FullMethod)
// 		ua := userAgent(ss.Context())

// 		streamLabels := prometheus.Labels{
// 			"service":      service,
// 			"handler":      method,
// 			userAgentLabel: ua,
// 		}
// 		if i.trackPeers {
// 			if ss != nil {
// 				streamLabels["peer"] = peerValue(ss.Context())
// 			} else {
// 				// mostly for testing purposes
// 				streamLabels["peer"] = "nil-server-stream"
// 			}
// 		}
// 		err := handler(srv, &monitoredServerStream{ServerStream: ss, labels: streamLabels, monitor: monitor})
// 		code := grpc.Code(err)
// 		labels := prometheus.Labels{
// 			"service":      service,
// 			"handler":      method,
// 			"code":         code.String(),
// 			"type":         handlerType(info.IsClientStream, info.IsServerStream),
// 			userAgentLabel: ua,
// 		}
// 		if i.trackPeers {
// 			if ss != nil {
// 				labels["peer"] = peerValue(ss.Context())
// 			} else {
// 				// mostly for testing purposes
// 				labels["peer"] = "nil-server-stream"
// 			}
// 		}
// 		if err != nil && code != codes.OK {
// 			monitor.errors.With(labels).Add(1)
// 		}

// 		monitor.requestDuration.With(labels).Observe(time.Since(start).Seconds())
// 		monitor.requestsTotal.With(labels).Add(1)

// 		return err
// 	}
// }

// // Describe implements prometheus Collector interface.
// func (i *Interceptor) Describe(in chan<- *prometheus.Desc) {
// 	i.monitoring.dialer.Describe(in)
// 	i.monitoring.server.Describe(in)
// 	i.monitoring.client.Describe(in)
// }

// // Collect implements prometheus Collector interface.
// func (i *Interceptor) Collect(in chan<- prometheus.Metric) {
// 	i.monitoring.dialer.Collect(in)
// 	i.monitoring.server.Collect(in)
// 	i.monitoring.client.Collect(in)
// }

// type ctxKey int

// var (
// 	tagRPCKey  ctxKey = 1
// 	tagConnKey ctxKey = 2
// )

// // TagRPC implements stats Handler interface.
// func (i *Interceptor) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
// 	service, method := split(info.FullMethodName)

// 	return context.WithValue(ctx, tagRPCKey, prometheus.Labels{
// 		"fail_fast": strconv.FormatBool(info.FailFast),
// 		"service":   service,
// 		"handler":   method,
// 	})
// }

// // HandleRPC implements stats Handler interface.
// func (i *Interceptor) HandleRPC(ctx context.Context, stat stats.RPCStats) {
// 	lab, _ := ctx.Value(tagRPCKey).(prometheus.Labels)

// 	switch in := stat.(type) {
// 	case *stats.Begin:
// 		if in.IsClient() {
// 			i.monitoring.client.requests.With(lab).Inc()
// 		} else {
// 			i.monitoring.server.requests.With(withUserAgentLabel(ctx, lab)).Inc()
// 		}
// 	case *stats.End:
// 		if in.IsClient() {
// 			i.monitoring.client.requests.With(lab).Dec()
// 		} else {
// 			i.monitoring.server.requests.With(withUserAgentLabel(ctx, lab)).Dec()
// 		}
// 	}
// }

// // TagConn implements stats Handler interface.
// func (i *Interceptor) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
// 	return context.WithValue(ctx, tagConnKey, prometheus.Labels{
// 		"remote_addr": info.RemoteAddr.String(),
// 		"local_addr":  info.LocalAddr.String(),
// 	})
// }

// // HandleConn implements stats Handler interface.
// func (i *Interceptor) HandleConn(ctx context.Context, stat stats.ConnStats) {
// 	lab, _ := ctx.Value(tagConnKey).(prometheus.Labels)

// 	switch in := stat.(type) {
// 	case *stats.ConnBegin:
// 		if in.IsClient() {
// 			i.monitoring.client.connections.With(lab).Inc()
// 		} else {
// 			i.monitoring.server.connections.With(withUserAgentLabel(ctx, lab)).Inc()
// 		}
// 	case *stats.ConnEnd:
// 		if in.IsClient() {
// 			i.monitoring.client.connections.With(lab).Dec()
// 		} else {
// 			i.monitoring.server.connections.With(withUserAgentLabel(ctx, lab)).Dec()
// 		}
// 	}
// }

// type monitoring struct {
// 	dialer *prometheus.CounterVec
// 	server *monitor
// 	client *monitor
// }

// type monitor struct {
// 	connections      *prometheus.GaugeVec
// 	requests         *prometheus.GaugeVec
// 	requestsTotal    *prometheus.CounterVec
// 	requestDuration  *prometheus.HistogramVec
// 	messagesReceived *prometheus.CounterVec
// 	messagesSend     *prometheus.CounterVec
// 	errors           *prometheus.CounterVec
// }

// // Describe implements prometheus Collector interface.
// func (m *monitor) Describe(in chan<- *prometheus.Desc) {
// 	// Gauge
// 	m.connections.Describe(in)
// 	m.requests.Describe(in)

// 	// HistogramVec
// 	m.requestDuration.Describe(in)

// 	// CounterVec
// 	m.requestsTotal.Describe(in)
// 	m.messagesReceived.Describe(in)
// 	m.messagesSend.Describe(in)
// 	m.errors.Describe(in)
// }

// // Collect implements prometheus Collector interface.
// func (m *monitor) Collect(in chan<- prometheus.Metric) {
// 	// Gauge
// 	m.connections.Collect(in)
// 	m.requests.Collect(in)

// 	// HistogramVec
// 	m.requestDuration.Collect(in)

// 	// CounterVec
// 	m.requestsTotal.Collect(in)
// 	m.messagesReceived.Collect(in)
// 	m.messagesSend.Collect(in)
// 	m.errors.Collect(in)
// }

// func initMonitoring(trackPeers bool, constLabels prometheus.Labels) *monitoring {
// 	dialer := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "reconnects_total",
// 			Help:        "Total number of reconnects made by client.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"address"},
// 	)

// 	serverConnections := prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "connections",
// 			Help:        "Number of currently opened server side connections.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"remote_addr", "local_addr", userAgentLabel},
// 	)
// 	serverRequests := prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "requests",
// 			Help:        "Number of currently processed server side rpc requests.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"fail_fast", "handler", "service", userAgentLabel},
// 	)
// 	serverRequestsTotal := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "requests_total",
// 			Help:        "Total number of RPC requests received by server.",
// 			ConstLabels: constLabels,
// 		},
// 		appendIf(trackPeers, []string{"service", "handler", "code", "type", userAgentLabel}, "peer"),
// 	)
// 	serverReceivedMessages := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "received_messages_total",
// 			Help:        "Total number of RPC messages received by server.",
// 			ConstLabels: constLabels,
// 		},
// 		appendIf(trackPeers, []string{"service", "handler", userAgentLabel}, "peer"),
// 	)
// 	serverSendMessages := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "send_messages_total",
// 			Help:        "Total number of RPC messages send by server.",
// 			ConstLabels: constLabels,
// 		},
// 		appendIf(trackPeers, []string{"service", "handler", userAgentLabel}, "peer"),
// 	)
// 	serverRequestDuration := prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "request_duration_seconds",
// 			Help:        "The RPC request latencies in seconds on server side.",
// 			ConstLabels: constLabels,
// 		},
// 		appendIf(trackPeers, []string{"service", "handler", "code", "type", userAgentLabel}, "peer"),
// 	)
// 	serverErrors := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "server",
// 			Name:        "errors_total",
// 			Help:        "Total number of errors that happen during RPC calles on server side.",
// 			ConstLabels: constLabels,
// 		},
// 		appendIf(trackPeers, []string{"service", "handler", "code", "type", userAgentLabel}, "peer"),
// 	)

// 	clientConnections := prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "connections",
// 			Help:        "Number of currently opened client side connections.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"remote_addr", "local_addr"},
// 	)
// 	clientRequests := prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "requests",
// 			Help:        "Number of currently processed client side rpc requests.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"fail_fast", "handler", "service"},
// 	)
// 	clientRequestsTotal := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "requests_total",
// 			Help:        "Total number of RPC requests made by client.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"service", "handler", "code", "type"},
// 	)
// 	clientReceivedMessages := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "received_messages_total",
// 			Help:        "Total number of RPC messages received.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"service", "handler"},
// 	)
// 	clientSendMessages := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "send_messages_total",
// 			Help:        "Total number of RPC messages send.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"service", "handler"},
// 	)
// 	clientRequestDuration := prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "request_duration_seconds",
// 			Help:        "The RPC request latencies in seconds on client side.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"service", "handler", "code", "type"},
// 	)
// 	clientErrors := prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace:   "grpc",
// 			Subsystem:   "client",
// 			Name:        "errors_total",
// 			Help:        "Total number of errors that happen during RPC calls.",
// 			ConstLabels: constLabels,
// 		},
// 		[]string{"service", "handler", "code", "type"},
// 	)

// 	return &monitoring{
// 		dialer: dialer,
// 		server: &monitor{
// 			connections:      serverConnections,
// 			requests:         serverRequests,
// 			requestsTotal:    serverRequestsTotal,
// 			requestDuration:  serverRequestDuration,
// 			messagesReceived: serverReceivedMessages,
// 			messagesSend:     serverSendMessages,
// 			errors:           serverErrors,
// 		},
// 		client: &monitor{
// 			connections:      clientConnections,
// 			requests:         clientRequests,
// 			requestsTotal:    clientRequestsTotal,
// 			requestDuration:  clientRequestDuration,
// 			messagesReceived: clientReceivedMessages,
// 			messagesSend:     clientSendMessages,
// 			errors:           clientErrors,
// 		},
// 	}
// }

// type monitoredServerStream struct {
// 	grpc.ServerStream
// 	labels  prometheus.Labels
// 	monitor *monitor
// }

// func (mss *monitoredServerStream) SendMsg(m interface{}) error {
// 	err := mss.ServerStream.SendMsg(m)
// 	if err == nil {
// 		mss.monitor.messagesSend.With(mss.labels).Inc()
// 	}
// 	return err
// }

// func (mss *monitoredServerStream) RecvMsg(m interface{}) error {
// 	err := mss.ServerStream.RecvMsg(m)
// 	if err == nil {
// 		mss.monitor.messagesReceived.With(mss.labels).Inc()
// 	}
// 	return err
// }

// type monitoredClientStream struct {
// 	grpc.ClientStream
// 	labels  prometheus.Labels
// 	monitor *monitor
// }

// func (mcs *monitoredClientStream) SendMsg(m interface{}) error {
// 	err := mcs.ClientStream.SendMsg(m)
// 	if err == nil {
// 		mcs.monitor.messagesSend.With(mcs.labels).Inc()
// 	}
// 	return err
// }

// func (mcs *monitoredClientStream) RecvMsg(m interface{}) error {
// 	err := mcs.ClientStream.RecvMsg(m)
// 	if err == nil {
// 		mcs.monitor.messagesReceived.With(mcs.labels).Inc()
// 	}
// 	return err
// }

// func handlerType(clientStream, serverStream bool) string {
// 	switch {
// 	case !clientStream && !serverStream:
// 		return "unary"
// 	case !clientStream && serverStream:
// 		return "server_stream"
// 	case clientStream && !serverStream:
// 		return "client_stream"
// 	default:
// 		return "bidirectional_stream"
// 	}
// }

// func split(name string) (string, string) {
// 	if i := strings.LastIndex(name, "/"); i >= 0 {
// 		return name[1:i], name[i+1:]
// 	}
// 	return "unknown", "unknown"
// }

// func userAgent(ctx context.Context) string {
// 	if md, ok := metadata.FromIncomingContext(ctx); ok {
// 		if ua, ok := md["user-agent"]; ok {
// 			return ua[0]
// 		}
// 	}
// 	return "not-set"
// }

// func withUserAgentLabel(ctx context.Context, lab prometheus.Labels) prometheus.Labels {
// 	lab[userAgentLabel] = userAgent(ctx)
// 	return lab
// }

// func peerValue(ctx context.Context) string {
// 	v, ok := peer.FromContext(ctx)
// 	if !ok {
// 		return "none"
// 	}
// 	return v.Addr.String()
// }

// func appendIf(ok bool, arr []string, val string) []string {
// 	if !ok {
// 		return arr
// 	}
// 	return append(arr, val)
// }
