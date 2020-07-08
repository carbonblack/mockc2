package c2

import (
	"encoding/hex"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

// A Generic protocol handler simply logs information about connections and
// the data received.
type Generic struct {
	delegate ProtocolDelegate
}

// SetDelegate saves the delegate for later use.
func (g *Generic) SetDelegate(delegate ProtocolDelegate) {
	g.delegate = delegate
}

// ReceiveData just logs information about data received.
func (g *Generic) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))
}
