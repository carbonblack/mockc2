package generic

import (
	"encoding/hex"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

// A Handler represents a generic protocol handler that simply logs information
// about connections and the data received.
type Handler struct {
	delegate protocol.Delegate
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// Accept gives the Handler a chance to do something as soon as an agent
// connects.
func (h *Handler) Accept() {
}

// ReceiveData just logs information about data received.
func (h *Handler) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	h.delegate.AgentConnected("")
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	log.Warn("generic doesn't support command execution")
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	log.Warn("generic doesn't support file upload")
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	log.Warn("generic doesn't support file download")
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}
