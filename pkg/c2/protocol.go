package c2

import (
	"net"
)

var protocols map[string]ProtocolHandler

func init() {
	protocols = make(map[string]ProtocolHandler)

	protocols["generic"] = Generic{}
}

// A ProtocolHandler represents a type capable of handling and decoding C2 traffic
type ProtocolHandler interface {
	ValidateConnection(conn net.Conn, quit chan interface{}) (*Agent, error)
	HandleConnection(conn net.Conn, quit chan interface{})
}

// ProtocolHandlerByName retrieves an instance of a specific protocol handler
func ProtocolHandlerByName(name string) ProtocolHandler {
	return protocols[name]
}

// ProtocolNames returns a list of names of protocol handlers
func ProtocolNames() []string {
	names := make([]string, len(protocols))
	i := 0
	for k := range protocols {
		names[i] = k
		i++
	}

	return names
}
