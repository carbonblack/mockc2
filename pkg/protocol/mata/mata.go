package mata

import (
	"bytes"
	"crypto/rc4"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"unicode/utf16"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/internal/queue"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
)

const (
	stateWaitingForBeacon1 = iota
	stateWaitingForBeacon3
	stateBeaconReceived
	stateWaitingForKeyLength
	stateWaitingForKey
	stateHandshakeComplete
)

const (
	opNone           uint32 = 0x00000000
	opBeacon1               = 0x00020000
	opBeacon2               = 0x00020100
	opBeacon3               = 0x00020200
	opSendRC4               = 0x00020300
	opSuccess               = 0x00020500
	opFailure               = 0x00020600
	opHostInfo              = 0x00000700
	opExecute               = 0x00010000
	opReverseExecute        = 0x00010002
	opFileUpload            = 0x00010100
	opFileDownload          = 0x00010101
	opFileDelete            = 0x00010103
	opFileScanDir           = 0x00010104
	opFileURLGet            = 0x00010110
)

// Handler is a Mata protocol handler.
type Handler struct {
	delegate       protocol.Delegate
	state          int
	keyLength      uint32
	key            []byte
	dataChan       chan byte
	sendCipher     *rc4.Cipher
	recvCipher     *rc4.Cipher
	pendingCommand uint32
	activeCommand  uint32
	file           *os.File
	blockCounter   int
	uploadFinished bool
}

type command struct {
	opcode  uint32
	size    uint32
	unknown uint32
	data    []byte
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// ReceiveData just logs information about data receiveh.
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
		switch h.state {
		case stateWaitingForBeacon1:
			b, err := queue.Get(h.dataChan, 4)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			opcode := binary.LittleEndian.Uint32(b)
			if opcode != opBeacon1 {
				h.delegate.CloseConnection()
				return
			}

			h.sendOpcode(opBeacon2)

			h.state = stateWaitingForBeacon3
		case stateWaitingForBeacon3:
			b, err := queue.Get(h.dataChan, 4)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			opcode := binary.LittleEndian.Uint32(b)

			if opcode != opBeacon3 {
				h.delegate.CloseConnection()
				return
			}

			h.state = stateBeaconReceived
		case stateBeaconReceived:
			cmd, err := h.recvPacket()
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			if cmd.opcode != opSendRC4 {
				h.delegate.CloseConnection()
				return
			}

			h.state = stateWaitingForKeyLength
		case stateWaitingForKeyLength:
			// Read RC4 key length
			b, err := queue.Get(h.dataChan, 4)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			h.keyLength = binary.LittleEndian.Uint32(b)
			h.key = make([]byte, h.keyLength)

			h.state = stateWaitingForKey
		case stateWaitingForKey:
			// Read RC4 key
			b, err := queue.Get(h.dataChan, int(h.keyLength))
			if err != nil {
				h.delegate.CloseConnection()
				return
			}
			copy(h.key, b)

			h.sendCipher, err = rc4.NewCipher(h.key)
			if err != nil {
				log.Warn("mata rc4 error: %v", err)
				h.delegate.CloseConnection()
				return
			}

			h.recvCipher, err = rc4.NewCipher(h.key)
			if err != nil {
				log.Warn("mata rc4 error: %v", err)
				h.delegate.CloseConnection()
				return
			}

			h.state = stateHandshakeComplete

			// Request host info
			h.sendPacket(command{
				opcode: opHostInfo,
			})
		case stateHandshakeComplete:
			// Receive data normally
			cmd, err := h.recvPacket()
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			h.processCommand(cmd)
		}

		h.delegate.AgentConnected("")
	}
}

func (h *Handler) sendOpcode(opcode uint32) {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, opcode)
	h.delegate.SendData(out)
}

func (h *Handler) sendPacket(cmd command) {
	log.Debug("sent\n")
	h.logCommand(cmd)

	out := make([]byte, 12)

	binary.LittleEndian.PutUint32(out[0:4], cmd.opcode)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(cmd.data)))
	binary.LittleEndian.PutUint32(out[8:12], cmd.unknown)

	if h.state == stateHandshakeComplete {
		encrypted := make([]byte, 12)
		h.sendCipher.XORKeyStream(encrypted, out)
		out = encrypted
	}

	h.delegate.SendData(out)

	if len(cmd.data) > 0 {
		data := cmd.data

		if h.state == stateHandshakeComplete {
			encrypted := make([]byte, len(cmd.data))
			h.sendCipher.XORKeyStream(encrypted, data)
			data = encrypted
		}

		h.delegate.SendData(data)
	}
}

func (h *Handler) recvPacket() (command, error) {
	cmd := command{}

	b, err := queue.Get(h.dataChan, 12)
	if err != nil {
		return cmd, err
	}

	if h.state == stateHandshakeComplete {
		encrypted := make([]byte, 12)
		h.recvCipher.XORKeyStream(encrypted, b)
		b = encrypted
	}

	cmd.opcode = binary.LittleEndian.Uint32(b[0:4])
	cmd.size = binary.LittleEndian.Uint32(b[4:8])
	cmd.unknown = binary.LittleEndian.Uint32(b[8:12])

	if cmd.size > 0 {
		cmd.data = make([]byte, cmd.size)

		data, err := queue.Get(h.dataChan, int(cmd.size))
		if err != nil {
			return cmd, err
		}

		if h.state == stateHandshakeComplete {
			encrypted := make([]byte, len(cmd.data))
			h.recvCipher.XORKeyStream(encrypted, data)
			data = encrypted
		}

		copy(cmd.data, data)
	}

	return cmd, nil
}

func (h *Handler) processCommand(cmd command) {
	log.Debug("received\n")
	h.logCommand(cmd)

	switch h.activeCommand {
	case opExecute:
		h.processExecute(cmd)
	case opFileDownload:
		h.processDownload(cmd)
	case opFileUpload:
		h.processUpload(cmd)
	default:
		switch cmd.opcode {
		case opSuccess:
			h.activeCommand = h.pendingCommand
			h.pendingCommand = opNone
			log.Info("mata command acknowledged")
		case opFailure:
			h.pendingCommand = opNone
			log.Warn("mata command failed")
		}
	}
}

func (h *Handler) processExecute(cmd command) {
	if cmd.unknown == 0x2 {
		log.Info(decodeWideString(cmd.data))
	} else if cmd.unknown == 0x1 {
		h.activeCommand = opNone
		log.Success("mata command succeeded")
	}
}

func (h *Handler) processDownload(cmd command) {
	if cmd.unknown == 0x0 {
		if cmd.size == 4 {
			// Send the file offset to the server
			fileOffset := uint32(0x0)
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, fileOffset)

			h.sendPacket(command{
				opcode:  opSuccess,
				size:    uint32(len(buf)),
				unknown: 0x0,
				data:    buf,
			})
		} else if cmd.size == 8 {
			// Ignore the file modification timestamp
		}
	} else if cmd.unknown == 0x2 {
		h.file.Write(cmd.data)

		h.blockCounter++
		if h.blockCounter == 16 {
			h.sendPacket(command{
				opcode: opSuccess,
			})
			h.blockCounter = 0
		}
	} else if cmd.unknown == 0x1 {
		h.file.Close()
		h.file = nil
		h.activeCommand = opNone
		h.blockCounter = 0

		h.sendPacket(command{
			opcode: opSuccess,
		})
		log.Success("mata command succeeded")
	}
}

func (h *Handler) processUpload(cmd command) {
	if cmd.unknown == 0x0 {
		if cmd.size == 0 {
			if h.uploadFinished {
				// Client acknowledged our end of file we can close our end
				h.file.Close()
				h.file = nil
				h.activeCommand = opNone
				h.blockCounter = 0

				log.Success("mata command succeeded")
			} else {
				h.sendFileChunk()
			}
		}
		if cmd.size == 4 {
			// Send the file offset and file to the server
			fileOffset := uint32(0x0)
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, fileOffset)

			h.sendPacket(command{
				opcode:  opSuccess,
				size:    uint32(len(buf)),
				unknown: 0x0,
				data:    buf,
			})

			h.sendFileChunk()
		}
	}
}

func (h *Handler) sendFileChunk() {
	buf := make([]byte, 0x4000)
	for {
		bytesRead, err := h.file.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Warn("Error reading source file; %v", err)
			}
			h.uploadFinished = true
			break
		}

		h.sendPacket(command{
			opcode:  opSuccess,
			size:    uint32(bytesRead),
			unknown: 0x2,
			data:    buf[:bytesRead],
		})

		h.blockCounter++
		if h.blockCounter == 16 {
			// Break so client can ack
			h.blockCounter = 0
			break
		}
	}

	if h.uploadFinished {
		h.sendPacket(command{
			opcode:  opSuccess,
			unknown: 0x1,
		})
	}
}

func (h *Handler) logCommand(cmd command) {
	log.Debug("Mata Command")
	log.Debug(" Opcode: 0x%08x", cmd.opcode)
	log.Debug("   Size: 0x%08x", cmd.size)
	log.Debug("Unknown: 0x%08x", cmd.unknown)
	if cmd.data != nil {
		log.Debug("   Data:\n%s", hex.Dump(cmd.data))
	}
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	commandLine := encodeWideString(strings.TrimSpace(name + " " + strings.Join(args, " ")))

	h.sendPacket(command{
		opcode:  opExecute,
		size:    uint32(len(commandLine) + 2),
		unknown: 0x0,
		data:    append([]byte(commandLine), 0x00, 0x00),
	})
	h.pendingCommand = opExecute
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	file, err := os.Open(source)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}

	h.file = file

	destWide := encodeWideString(destination)

	h.sendPacket(command{
		opcode:  opFileUpload,
		size:    uint32(len(destWide) + 2),
		unknown: 0x0,
		data:    append([]byte(destWide), 0x00, 0x00),
	})
	h.pendingCommand = opFileUpload
	h.uploadFinished = false
	h.blockCounter = 0
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	file, err := os.Create(destination)
	if err != nil {
		log.Warn("Error opening destination file: %v", err)
		return
	}

	h.file = file

	sourceWide := encodeWideString(source)

	h.sendPacket(command{
		opcode:  opFileDownload,
		size:    uint32(len(sourceWide) + 2),
		unknown: 0x0,
		data:    append([]byte(sourceWide), 0x00, 0x00),
	})
	h.pendingCommand = opFileDownload
	h.blockCounter = 0
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
	if h.file != nil {
		h.file.Close()
		h.file = nil
	}
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return true
}

func encodeWideString(input string) []byte {
	buff := new(bytes.Buffer)
	ws := utf16.Encode([]rune(input))

	for _, c := range ws {
		binary.Write(buff, binary.LittleEndian, c)
	}

	return buff.Bytes()
}

func decodeWideString(input []byte) string {
	ws := make([]uint16, 0)

	for i := range input {
		if i%2 == 1 {
			ws = append(ws, binary.LittleEndian.Uint16([]byte{input[i-1], input[i]}))
		}
	}

	return string(utf16.Decode(ws))
}
