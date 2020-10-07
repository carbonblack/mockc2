package rifdoor

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/internal/queue"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

const (
	beacon   uint32 = 0x9e2
	request  uint32 = 0x4e3a
	response uint32 = 0xa021
	end      uint32 = 0x1055
)

// Handler is a Rifdoor protocol handler capable of communicating with the
// Rifdoor malware family.
type Handler struct {
	delegate protocol.Delegate
	dataChan chan byte
	checksum uint32
}

type command struct {
	opcode   uint32
	checksum uint32
	zero     uint32
	size     uint32
	data     []byte
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// Accept gives the Handler a chance to do something as soon as an agent
// connects.
func (h *Handler) Accept() {
}

// ReceiveData saves the network data and processes it when a full command has
// been received.
func (h *Handler) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	if h.dataChan == nil {
		h.dataChan = make(chan byte, len(data))
		go h.processData()
	}

	queue.Put(h.dataChan, data)
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	commandLine := strings.TrimSpace(name + " " + strings.Join(args, " "))

	rc := command{
		opcode:   request,
		checksum: h.checksum,
		zero:     0x0,
		size:     uint32(len(commandLine)),
		data:     []byte(commandLine),
	}
	data := h.encodeCommand(rc)
	h.delegate.SendData(data)

	rc = command{
		opcode:   end,
		checksum: h.checksum,
		zero:     0x0,
		size:     0x0,
	}
	data = h.encodeCommand(rc)
	h.delegate.SendData(data)
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	log.Warn("rifdoor doesn't support file upload")
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	log.Warn("rifdoor doesn't support file download")
}

func (h *Handler) encodeCommand(cmd command) []byte {
	result := make([]byte, 16+cmd.size)

	binary.LittleEndian.PutUint32(result[0:], cmd.opcode)
	binary.LittleEndian.PutUint32(result[4:], cmd.checksum)
	binary.LittleEndian.PutUint32(result[8:], cmd.zero)
	binary.LittleEndian.PutUint32(result[12:], cmd.size)

	if cmd.size > 0 {
		copy(result[16:], cipher(cmd.data))
	}

	return result
}

// Close cleans up any uzed resources.
func (h *Handler) Close() {
	close(h.dataChan)
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}

func (h *Handler) processData() {
	for {
		cmd := command{}

		// Receive the header
		b, err := queue.Get(h.dataChan, 16)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		cmd.opcode = binary.LittleEndian.Uint32(b[0:4])
		cmd.checksum = binary.LittleEndian.Uint32(b[4:8])
		cmd.zero = binary.LittleEndian.Uint32(b[8:12])
		cmd.size = binary.LittleEndian.Uint32(b[12:16])

		if cmd.size > 0 {
			b, err = queue.Get(h.dataChan, int(cmd.size))
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			cmd.data = cipher(b)
		}

		h.processCommand(cmd)
	}
}

func (h *Handler) processCommand(cmd command) {
	h.logCommand(cmd)

	h.checksum = cmd.checksum

	switch cmd.opcode {
	case beacon:
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, cmd.checksum)
		hash := sha256.Sum256(data)
		id := hex.EncodeToString(hash[:])

		h.delegate.AgentConnected(id)
	case response:
		log.Info(string(cmd.data))
	case end:
		h.delegate.CloseConnection()
	}
}

func (h *Handler) logCommand(cmd command) {
	log.Debug("Rifdoor Command")
	log.Debug("  Opcode: 0x%08x", cmd.opcode)
	log.Debug("Checksum: 0x%08x", cmd.checksum)
	log.Debug("    Zero: 0x%08x", cmd.zero)
	log.Debug("    Size: 0x%08x", cmd.size)
	if cmd.data != nil {
		log.Debug("    Data:\n%s", hex.Dump(cmd.data))
	}
}

func byte1(i int) int {
	return (i & 0x0000FF00) >> 8
}

func byte2(i int) int {
	return (i & 0x00FF0000) >> 16
}

func hibyte(i int) int {
	return (i & 0xFF000000) >> 24
}

func cipher(input []byte) []byte {
	output := make([]byte, len(input))

	key1 := 0x1A2C
	key2 := 0x1A2C
	key3 := 0x4C5B

	for i := 0; i < len(input); i++ {
		v6 := (key3 ^ key2&byte1(key1) ^ int(input[i]) ^ byte2(key1)&hibyte(key1) ^ byte1(key3)&byte2(key3)&hibyte(key3)) & 0xff
		v7 := ((key3 >> 8) | (key2 << 24)) & 0xffffffff
		key1 = ((key1 >> 8) | (((16 * key3) ^ (key3^(2*(key3^(4*key3))))&0xFFFFFFF0) << 20)) & 0xffffffff
		output[i] = byte(v6 & 0xff)
		key2 = key1
		key3 = v7
	}

	return output
}
