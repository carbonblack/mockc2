package c2

import (
	"fmt"
)

// A ProtocolDelegate represents an external delegate that the protocolhandler
// can notify about data being processed.
type ProtocolDelegate interface {
	SendData(data []byte)
	AgentConnected(a *Agent)
	CloseConnection()
}

// A ProtocolHandler represents a type capable of handling and decoding C2
// traffic.
type ProtocolHandler interface {
	SetDelegate(delegate ProtocolDelegate)
	ReceiveData(data []byte)
	Close()
}

// NewProtocolHandler creates a concrete instance of a given protocol handler.
func NewProtocolHandler(protocol string) (ProtocolHandler, error) {
	switch protocol {
	case "generic":
		return &Generic{}, nil
	case "hotcroissant":
		return &HotCroissant{}, nil
	case "rifdoor":
		return &Rifdoor{}, nil
	default:
		return nil, fmt.Errorf("unknown protocol %s", protocol)
	}
}
