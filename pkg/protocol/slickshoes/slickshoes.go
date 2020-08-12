package slickshoes

import (
	"bytes"
	"crypto/sha256"
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
	opBeacon                uint32 = 0x00000000
	opUninstall             uint32 = 0x00010000
	opShutdown              uint32 = 0x00020000
	opDirectoryGet          uint32 = 0x00000001
	opDirectorySet          uint32 = 0x00010001
	opExecute               uint32 = 0x00020001
	opExecuteStop           uint32 = 0x00030001
	opFileList              uint32 = 0x00000002
	opFileFolderList        uint32 = 0x00010002
	opFileDownload          uint32 = 0x00020002
	opFileError             uint32 = 0x00020010
	opFileUpload            uint32 = 0x00030002
	opScreenCaptureStart    uint32 = 0x00000003
	opScreenCaptureStop     uint32 = 0x00010003
	opScreenCaptureInterval uint32 = 0x00020003
)

// Handler is a Slickshoes protocol handler capable of communicating with the
// Slickshoes malware family.
type Handler struct {
	delegate protocol.Delegate
	dataChan chan byte
	fileName string
	file     *os.File
}

type command struct {
	size   uint32
	opcode uint32
	opt    uint16
	data   []byte
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

// Close cleans up any uzed resources.
func (h *Handler) Close() {
	close(h.dataChan)
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	commandLine := encodeWideString(strings.TrimSpace(name + " " + strings.Join(args, " ")))
	data := append(commandLine, []byte{0x00, 0x00}...)

	h.sendData(command{
		size:   uint32(len(data)),
		opcode: opExecute,
		opt:    0x0001,
		data:   data,
	})
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	file, err := os.Open(source)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}
	defer file.Close()

	ws := append(encodeWideString(destination), []byte{0x00, 0x00}...)
	h.sendData(command{
		size:   uint32(len(ws)),
		opcode: opFileUpload,
		opt:    0x0001,
		data:   ws,
	})

	buf := make([]byte, 0x40000)
	for {
		bytesRead, err := file.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Warn("Error reading source file; %v", err)
			}

			break
		}

		h.sendData(command{
			size:   uint32(bytesRead),
			opcode: opFileUpload,
			opt:    0x0000,
			data:   buf[:bytesRead],
		})
	}

	// Finish the file transfer
	h.sendData(command{
		size:   0x00000000,
		opcode: opFileUpload,
		opt:    0x0100,
	})
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	h.fileName = destination

	ws := append(encodeWideString(destination), []byte{0x00, 0x00}...)
	h.sendData(command{
		size:   uint32(len(ws)),
		opcode: opFileDownload,
		opt:    0x0001,
		data:   ws,
	})
}

func (h *Handler) processData() {
	for {
		cmd := command{}

		b, err := queue.Get(h.dataChan, 10)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		cmd.size = binary.LittleEndian.Uint32(b[0:4])
		cmd.opcode = binary.LittleEndian.Uint32(b[4:8])
		cmd.opt = binary.LittleEndian.Uint16(b[8:10])

		if cmd.size > 0 {
			cmd.data = make([]byte, cmd.size)

			data, err := queue.Get(h.dataChan, int(cmd.size))
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			data = cipher(data)

			copy(cmd.data, data)
		}

		h.proccessCommand(cmd)
	}
}

func (h *Handler) proccessCommand(cmd command) {
	logCommand(cmd)

	switch cmd.opcode {
	case opBeacon:
		if cmd.size == 0x88 {
			hash := sha256.Sum256(cmd.data)
			id := hex.EncodeToString(hash[:])

			h.delegate.AgentConnected(id)
		}
	case opExecute:
		log.Info(decodeWideString(cmd.data))
		if cmd.opt == 0x0100 {
			log.Success("Execute complete")
		}
	case opFileError:
		if h.file != nil {
			h.file.Close()
			h.file = nil
			h.fileName = ""
		}
		log.Warn("Error transferring file")
	case opFileUpload:
		if cmd.opt == 0x0100 {
			log.Success("Upload complete")
		}
	case opFileDownload:
		switch cmd.opt {
		case 0x0001:
			file, err := os.Create(h.fileName)
			if err != nil {
				log.Warn("Error opening destination file: %v", err)
				return
			}
			h.file = file
		case 0x0000:
			if h.file != nil {
				h.file.Write(cmd.data)
			}
		case 0x0100:
			if h.file != nil {
				h.file.Close()
				h.file = nil
				h.fileName = ""
			}
			log.Success("Download complete")
		}
	}
}

func logCommand(c command) {
	log.Debug("Slickshoes Command")
	log.Debug("  Size: 0x%x", c.size)
	log.Debug("Opcode: 0x%x", c.opcode)
	log.Debug("   Opt: 0x%04x", c.opt)
	log.Debug("  Data:\n%s", hex.Dump(c.data))
}

func (h *Handler) sendData(cmd command) error {
	result := make([]byte, 10+len(cmd.data))

	binary.LittleEndian.PutUint32(result[0:4], cmd.size)
	binary.LittleEndian.PutUint32(result[4:8], cmd.opcode)
	binary.LittleEndian.PutUint16(result[8:10], cmd.opt)

	if cmd.size > 0 {
		encrypted := cipher(cmd.data)
		copy(result[10:], encrypted)
	}

	h.delegate.SendData(result)

	return nil
}

func cipher(input []byte) []byte {
	output := make([]byte, len(input))

	key1 := 0x49
	key2 := 0x1310a024
	key3 := 0xa323da32

	for i := 0; i < len(input); i++ {
		output[i] = byte((int(input[i]) ^ key3 ^ key1) & 0xff)
		tmp1 := key3 >> 8
		key1 = (key2>>0x10)&(key2>>8)&key2 ^ (key3>>0x10)&tmp1 ^ key3&key1 ^ (key3 >> 0x18)
		tmp2 := key3*2 ^ key3
		key3 = key2<<0x18 | key3>>8
		key2 = (tmp2&0x1fe)<<0x16 | key2>>8
	}

	return output
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
