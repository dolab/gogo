package listeners

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/dolab/gogo/pkgs/hooks"
)

// A Listener implments Interface
type Listener struct {
	http2  bool
	tlscfg *tls.Config

	mux     sync.RWMutex
	network string
	address string
	conn    io.Closer
}

func New(isHTTP2 bool) Interface {
	return &Listener{
		http2: isHTTP2,
	}
}

func (l *Listener) WithTLSConfig(cfg *tls.Config) Interface {
	l.mux.Lock()
	l.tlscfg = cfg
	l.mux.Unlock()

	return l
}

// Listen announces on the local network address.
//
// The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
//
// For TCP networks, if the host in the address parameter is empty or
// a literal unspecified IP address, Listen listens on all available
// unicast and anycast IP addresses of the local system.
// To only use IPv4, use network "tcp4".
// The address can use a host name, but this is not recommended,
// because it will create a listener for at most one of the host's IP
// addresses.
// If the port in the address parameter is empty or "0", as in
// "127.0.0.1:" or "[::1]:0", a port number is automatically chosen.
// The Addr method of Listener can be used to discover the chosen
// port.
//
// See func Dial for a description of the network and address
// parameters.
func (l *Listener) Listen(network, address string) (conn net.Listener, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.conn != nil {
		err = fmt.Errorf("%s:%s already in use", l.network, l.address)
		return
	}

	l.network = network
	l.address = address

	conn, err = net.Listen(network, address)
	if err == nil {
		l.conn = conn
	}
	return
}

// ListenPacket announces on the local network address.
//
// The network must be "udp", "udp4", "udp6", "unixgram", or an IP
// transport. The IP transports are "ip", "ip4", or "ip6" followed by
// a colon and a literal protocol number or a protocol name, as in
// "ip:1" or "ip:icmp".
//
// For UDP and IP networks, if the host in the address parameter is
// empty or a literal unspecified IP address, ListenPacket listens on
// all available IP addresses of the local system except multicast IP
// addresses.
// To only use IPv4, use network "udp4" or "ip4:proto".
// The address can use a host name, but this is not recommended,
// because it will create a listener for at most one of the host's IP
// addresses.
// If the port in the address parameter is empty or "0", as in
// "127.0.0.1:" or "[::1]:0", a port number is automatically chosen.
// The LocalAddr method of PacketConn can be used to discover the
// chosen port.
//
// See func Dial for a description of the network and address
// parameters.
func (l *Listener) ListenPacket(network, address string) (conn net.PacketConn, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.conn != nil {
		err = fmt.Errorf("%s:%s already in use", l.network, l.address)
		return
	}

	l.network = network
	l.address = address

	conn, err = net.ListenPacket(network, address)
	if err == nil {
		l.conn = conn
	}
	return
}

func (l *Listener) Shutdown() {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.conn == nil {
		return
	}

	if err := l.conn.Close(); err == nil {
		l.conn = nil
	}

	return
}

func (l *Listener) RequestReceivedHook() hooks.NamedHook {
	return hooks.NamedHook{
		Name: "__listener@default",
		Apply: func(w http.ResponseWriter, r *http.Request) bool {
			switch l.network {
			case "unix":
				r.URL.Path = strings.TrimPrefix(r.URL.Path, l.address)
				r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, l.address)
			}

			return true
		},
	}
}
