package crosswalk

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/internal/queue"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

const (
	tlsHeaderLength = 5
)

const (
	opHandshakeAgent  uint32 = 0x00000065
	opHandshakeServer uint32 = 0x00000064
	opHostInfo        uint32 = 0x0000006f
	opHeartbeat       uint32 = 0x0000008d
)

type tlsHeader struct {
	contentType uint8
	version     uint16
	length      uint16
}

type command struct {
	tls    tlsHeader
	opcode uint32
	length uint32
	uuid   [36]byte
	data   []byte
}

// Handler is a Crosswalk protocol handler capable of communicating with the
// Crosswalk malware family.
type Handler struct {
	delegate          protocol.Delegate
	dataChan          chan byte
	serverUUID        string
	serverHash        [16]byte
	serverKey         [16]byte
	clientUUID        string
	clientHash        [16]byte
	clientKey         [16]byte
	handshakeComplete bool
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// Accept gives the Handler a chance to do something as soon as an agent
// connects.
func (h *Handler) Accept() {
	u, err := newUUID()
	if err != nil {
		log.Warn("Crosswalk error creating UUID")
	}

	h.serverUUID = u.String()
	h.serverHash = u.Hash()
	h.serverKey = generateKey(h.serverHash)
}

// ReceiveData just logs information about data received.
func (h *Handler) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	if h.dataChan == nil {
		h.dataChan = make(chan byte, len(data))
		go h.processData()
	}

	queue.Put(h.dataChan, data)
}

func (h *Handler) processData() {
	for {
		cmd := command{}

		b, err := queue.Get(h.dataChan, tlsHeaderLength)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		cmd.tls.contentType = b[0]
		cmd.tls.version = binary.BigEndian.Uint16(b[1:3])
		cmd.tls.length = binary.BigEndian.Uint16(b[3:5])

		data, err := queue.Get(h.dataChan, int(cmd.tls.length))
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		if h.handshakeComplete {
			data = aesDecrypt(data, h.serverKey[:])
		}

		cmd.opcode = binary.LittleEndian.Uint32(data[0:4])
		cmd.length = binary.LittleEndian.Uint32(data[4:8])
		copy(cmd.uuid[:], data[8:44])

		if cmd.length > 0 {
			cmd.data = data[44:]
		}

		h.proccessCommand(cmd)
	}
}

func (h *Handler) proccessCommand(cmd command) {
	logCommand(cmd)

	switch cmd.opcode {
	case opHandshakeAgent:
		s, _ := parseUUID(cmd.uuid[:])
		h.clientUUID = s

		hash := cmd.data[0:72]
		copy(h.clientHash[:], hash)
		h.clientKey = generateKey(h.clientHash)

		err := h.sendServerHandshake()
		if err != nil {
			log.Warn("Crosswalk error sending command: %v", err)
		}

		h.handshakeComplete = true
	case opHostInfo:
		hash := sha256.Sum256(cmd.data)
		id := hex.EncodeToString(hash[:])

		h.delegate.AgentConnected(id)
	}
}

func (h *Handler) sendServerHandshake() error {
	length := 72 - len(h.serverHash)
	hash := append(h.serverHash[:], bytes.Repeat([]byte{0x00}, length)...)

	tempKey := cryptDeriveKey(hash)
	hashCopy := make([]byte, len(hash))
	copy(hashCopy, hash)
	encryptedHash := aesEncrypt(hashCopy, tempKey)
	length = 144 - len(encryptedHash)
	encryptedHash = append(encryptedHash, bytes.Repeat([]byte{0x00}, length)...)

	data := make([]byte, len(hash)+len(encryptedHash))
	copy(data[0:72], hash)
	copy(data[72:], encryptedHash)

	return h.sendCommand(0x64, data)
}

// Shutdown network thread
func (h *Handler) sendCommandA0() error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 0x1)

	return h.sendCommand(0xa0, data)
}

// Execute or create
// 0x78 to stage
// 0x7a to upload data
// 0x78 to execute
func (h *Handler) sendCommand78() error {
	// 16 bytes
	// data[2] = code size
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:4], 0x1)
	binary.LittleEndian.PutUint32(data[4:8], 0x2)
	binary.LittleEndian.PutUint32(data[8:12], 0x8)  // code size
	binary.LittleEndian.PutUint32(data[12:16], 0x4) // something? plugin id

	// Put into struct in client
	// [0] = data[2] = code size
	// [1] = data[3] = ? plugin id maybe
	// [2] = 0 = initializd? offset? final size
	// [3] = VirtualAlloc result allocated buffer

	return h.sendCommand(0x78, data)
}

// Upload plugin
func (h *Handler) sendCommand7A() error {
	// 8 bytes
	// data[1] = code size
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:4], 0x4)
	binary.LittleEndian.PutUint32(data[4:8], 0x8)

	// The rest is data
	data[8] = 0xde
	data[9] = 0xed
	data[10] = 0xbe
	data[11] = 0xef

	return h.sendCommand(0x7a, data)
}

func logCommand(cmd command) {
	log.Debug("Crosswalk Command")
	log.Debug("TLS Content Type: %d", cmd.tls.contentType)
	log.Debug("     TLS Version: 0x%x", cmd.tls.version)
	log.Debug("      TLS Length: %d", cmd.tls.length)
	log.Debug("          Opcode: 0x%x", cmd.opcode)
	log.Debug("          Length: %d", cmd.length)
	log.Debug("            UUID:\n%s", hex.Dump(cmd.uuid[:]))
	if cmd.length > 0 {
		log.Debug("            Data:\n%s", hex.Dump(cmd.data))
	}
}

func (h *Handler) sendCommand(opcode uint32, data []byte) error {
	comm := command{}
	comm.tls.contentType = 0x17
	comm.tls.version = 0x301
	comm.tls.length = uint16(len(data) + 44)
	comm.opcode = opcode
	comm.length = uint32(len(data))

	uuid := []byte(h.serverUUID)
	length := 36 - len(uuid)
	uuid = append(uuid, bytes.Repeat([]byte{0x00}, length)...)
	copy(comm.uuid[:], uuid)

	comm.data = make([]byte, len(data))
	copy(comm.data, data)

	// Translate to wire format
	payload := make([]byte, comm.tls.length)
	binary.LittleEndian.PutUint32(payload[0:4], comm.opcode)
	binary.LittleEndian.PutUint32(payload[4:8], comm.length)
	copy(payload[8:44], comm.uuid[:])

	if len(data) > 0 {
		copy(payload[44:44+len(data)], data)
	}

	if h.handshakeComplete {
		payload = aesEncrypt(payload, h.clientKey[:])
	}

	// Update length since encrypting will pad to block size
	comm.tls.length = uint16(len(payload))

	buf := make([]byte, comm.tls.length+5)
	buf[0] = comm.tls.contentType
	binary.BigEndian.PutUint16(buf[1:3], comm.tls.version)
	binary.BigEndian.PutUint16(buf[3:5], comm.tls.length)
	copy(buf[5:], payload)

	h.delegate.SendData(buf)

	return nil
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
