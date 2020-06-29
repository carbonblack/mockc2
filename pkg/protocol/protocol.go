package protocol

import (
	"net"

	"megaman.genesis.local/sknight/mockc2/pkg/agents"
)

var protocols map[string]Handler

func init() {
	protocols = make(map[string]Handler)

	protocols["generic"] = Generic{}
}

// A Handler represents a type capable of handling and decoding C2 traffic
type Handler interface {
	ValidateConnection(conn net.Conn, quit chan interface{}) (*agents.Agent, error)
	HandleConnection(conn net.Conn, quit chan interface{})
}

// HandlerByName retrieves an instance of a specific protocol handler
func HandlerByName(name string) Handler {
	return protocols[name]
}

// Names returns a list of names of protocol handlers
func Names() []string {
	names := make([]string, len(protocols))
	i := 0
	for k := range protocols {
		names[i] = k
		i++
	}

	return names
}
