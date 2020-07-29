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

	a := &Agent{}
	g.delegate.AgentConnected(a)
}

// SendCommand sends a command to the connected agent.
func (g *Generic) SendCommand(command interface{}) {
	switch command.(type) {
	case ExecuteCommand:
		log.Warn("generic doesn't support command execution")
	case UploadCommand:
		log.Warn("generic doesn't support file upload")
	case DownloadCommand:
		log.Warn("generic doesn't support file download")
	}
}

// Close cleans up any uzed resources
func (g *Generic) Close() {
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (g *Generic) NeedsTLS() bool {
	return false
}
