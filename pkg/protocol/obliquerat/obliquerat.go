package obliquerat

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/internal/queue"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

const (
	opNone     = ""
	opDownload = "4"
	opExecute  = "7"
	opHostInfo = "0"
	opUpload   = "8"
)

// A Handler represents a ObliqueRat protocol handler that simply logs information
// about connections and the data received.
type Handler struct {
	delegate      protocol.Delegate
	dataChan      chan byte
	activeCommand string
	executable    string
	source        string
	destination   string
	file          *os.File
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// Accept gives the Handler a chance to do something as soon as an agent
// connects.
func (h *Handler) Accept() {
	if h.dataChan == nil {
		h.dataChan = make(chan byte, 4)
		go h.processData()
	}

	h.activeCommand = opHostInfo
	h.delegate.SendData([]byte(opHostInfo))
}

// ReceiveData just logs information about data received.
func (h *Handler) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	queue.Put(h.dataChan, data)
}

func (h *Handler) processData() {
	for {
		b, err := queue.Get(h.dataChan, 4)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}
		resp := string(b)

		if resp == "ack\x00" {
			log.Info("obliquerat command ack")
		} else if resp == "nak\x00" {
			log.Warn("obliquerat command nak")
			continue
		}

		switch h.activeCommand {
		case opDownload:
			h.processDownload()
		case opExecute:
			h.processExecute()
		case opHostInfo:
			h.processHostInfo()
		case opUpload:
			h.processUpload()
		}
	}
}

func (h *Handler) cleanup() {
	if h.file != nil {
		h.file.Close()
	}

	h.file = nil
	h.executable = ""
	h.source = ""
	h.destination = ""
	h.activeCommand = opNone
}

func (h *Handler) processHostInfo() {
	defer h.cleanup()

	var hostInfo bytes.Buffer
	for {
		b, err := queue.Get(h.dataChan, 1)
		if err != nil {
			h.delegate.CloseConnection()

			return
		}

		if b[0] == 0x00 {
			break
		}

		hostInfo.Write(b)
	}

	hash := sha256.Sum256(hostInfo.Bytes())
	id := hex.EncodeToString(hash[:])
	h.delegate.AgentConnected(id)
}

func (h *Handler) processExecute() {
	defer h.cleanup()

	data := append([]byte(h.executable), 0x00)
	h.delegate.SendData(data)
}

func (h *Handler) processDownload() {
	defer h.cleanup()

	data := append([]byte(h.source), 0x00)
	h.delegate.SendData(data)

	b, err := queue.Get(h.dataChan, 4)
	if err != nil {
		h.delegate.CloseConnection()
		return
	}
	resp := string(b)

	acked := false
	if resp == "ack\x00" {
		log.Info("obliquerat command ack")
		acked = true
	} else if resp == "nak\x00" {
		log.Warn("obliquerat command nak")
		acked = true
	}

	if acked {
		// There's not an ack after sending the requested file but it seems
		// like there should be so we handle it and read the next 4 bytes to
		// get the size. If there's no ack then we already have the size.
		b, err = queue.Get(h.dataChan, 4)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}
	}

	size := binary.BigEndian.Uint32(b)

	b, err = queue.Get(h.dataChan, int(size))
	if err != nil {
		h.delegate.CloseConnection()
		return
	}

	h.file.Write(b)
}

func (h *Handler) processUpload() {
	defer h.cleanup()

	data := append([]byte(h.destination), 0x00)
	h.delegate.SendData(data)

	b, err := queue.Get(h.dataChan, 4)
	if err != nil {
		h.delegate.CloseConnection()
		return
	}
	resp := string(b)

	if resp == "ack\x00" {
		log.Info("obliquerat command ack")
	} else if resp == "nak\x00" {
		log.Warn("obliquerat command nak")
		return
	}

	fi, err := h.file.Stat()
	if err != nil {
		return
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(fi.Size()))
	h.delegate.SendData(buf)

	buf = make([]byte, 0x4000)
	for {
		bytesRead, err := h.file.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Warn("Error reading source file; %v", err)
			}
			break
		}

		h.delegate.SendData(buf[:bytesRead])
	}
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	commandLine := strings.TrimSpace(name + " " + strings.Join(args, " "))

	h.executable = commandLine
	h.activeCommand = opExecute
	h.delegate.SendData([]byte(opExecute))
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	file, err := os.Open(source)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}

	h.file = file
	h.source = source
	h.destination = destination

	h.activeCommand = opUpload
	h.delegate.SendData([]byte(opUpload))
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	file, err := os.Create(destination)
	if err != nil {
		log.Warn("Error opening destination file: %v", err)
		return
	}

	h.file = file
	h.source = source
	h.destination = destination

	h.activeCommand = opDownload
	h.delegate.SendData([]byte(opDownload))
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}
