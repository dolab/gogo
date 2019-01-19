package listeners

import (
	"net"

	"github.com/dolab/gogo/pkgs/hooks"
)

// A Interface extends net.Listener with custom actions
type Interface interface {
	Listen(network, address string) (net.Listener, error)
	Shutdown()
	// Serve(server *http.Server)

	// hooks
	RequestReceivedHook() hooks.NamedHook
}
