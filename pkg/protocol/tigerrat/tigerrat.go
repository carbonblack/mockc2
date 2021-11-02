package tigerrat

import (
	"bytes"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/internal/queue"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

var keys = []string{
	"\xde\x42\xbe\x46\xea\xb9\xcd\xfc\x5c\xe3\x06\x64\x26\xc1\xfa\x1f\x73\x9f\x55\x74\x80\x96\x58\xf2\xad\x54\x8a\x57\xd4\x20\xaa\xb1",
}

var hashes = []string{
	"\xf2\x7c\x29\x1f\xa5\x75\xfa\x20\x23\xf7\x7b\x5b\xfa\x5b\xe1\x4a",
}

const (
	modMain          uint32 = 0x00000000
	modUpdate        uint32 = 0x00000001
	modInformation   uint32 = 0x00000002
	modShell         uint32 = 0x00000003
	modFileManager   uint32 = 0x00000004
	modKeyLogger     uint32 = 0x00000005
	modSocksTunnel   uint32 = 0x00000006
	modScreenCapture uint32 = 0x00000007
	modPortForwarder uint32 = 0x0000000a
)

const (
	opUpdateExit                      uint32 = 0x00000020
	opUpdateRemove                    uint32 = 0x00000030
	opInformationComputerName         uint32 = 0x00000010
	opInformationVersion              uint32 = 0x00000020
	opInformationAdapterInfo          uint32 = 0x00000030
	opInformationUserName             uint32 = 0x00000040
	opShellExecute                    uint32 = 0x00000010
	opShellSetDirectory               uint32 = 0x00000020
	opShellGetDirectory               uint32 = 0x00000030
	opShellSocket                     uint32 = 0x00000040
	opFileManagerListDrives           uint32 = 0x00000010
	opFileManagerListFiles            uint32 = 0x00000020
	opFileManagerFileDelete           uint32 = 0x00000030
	opFileManagerUploadStart          uint32 = 0x00000040
	opFileManagerUploadData           uint32 = 0x00000042
	opFileManagerUploadDone           uint32 = 0x00000043
	opFileManagerDownloadFile         uint32 = 0x00000050
	opFileManagerDownloadFilePosition uint32 = 0x00000057
	opFileManagerSetFlag              uint32 = 0x0000005f
	opFileManagerCreateProcess        uint32 = 0x00000060
	opFileManagerCreateProcessAsUser  uint32 = 0x00000063
	opFileManagerDownloadDirectory    uint32 = 0x00000070
	opFileManagerList1                uint32 = 0x00000080
	opFileManagerList2                uint32 = 0x00000090
)

// A Handler represents a TigerRAT protocol handler that simply logs information
// about connections and the data received.
type Handler struct {
	delegate          protocol.Delegate
	dataChan          chan byte
	hash              [16]byte
	sendCipher        *rc4.Cipher
	recvCipher        *rc4.Cipher
	handshakeComplete bool
	fileName          string
	file              *os.File
}

type command struct {
	module uint32
	opcode uint32
	size   uint32
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

// ReceiveData just logs information about data received.
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
	commandLine := encodeWideString(strings.TrimSpace(name + " " + strings.Join(args, " ")))
	data := append(commandLine, []byte{0x00, 0x00}...)

	h.sendCommand(command{
		module: modShell,
		opcode: opShellExecute,
		size:   uint32(len(data)),
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

	// The first four bytes are skipped
	ws := append([]byte{0x00, 0x00, 0x00, 0x00}, encodeWideString(destination)...)
	ws = append(ws, []byte{0x00, 0x00}...)
	h.sendCommand(command{
		module: modFileManager,
		opcode: opFileManagerUploadStart,
		size:   uint32(len(ws)),
		data:   ws,
	})

	h.file = file
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	h.fileName = destination

	ws := append(encodeWideString(source), []byte{0x00, 0x00}...)
	h.sendCommand(command{
		module: modFileManager,
		opcode: opFileManagerDownloadFile,
		size:   uint32(len(ws)),
		data:   ws,
	})
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}

func (h *Handler) processData() {
	for {
		if !h.handshakeComplete {
			b, err := queue.Get(h.dataChan, 44)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			h.delegate.SendData([]byte("HTTP 1.1 200 OK SSL2.1\x00"))

			b, err = queue.Get(h.dataChan, 17)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			copy(h.hash[:], b[0:16])

			key := ""
			for idx, val := range hashes {
				if val == string(h.hash[:]) {
					key = keys[idx]
				}
			}
			if key == "" {
				log.Warn("tigerrat unknown key")
				h.delegate.CloseConnection()
				return
			}

			h.recvCipher, err = rc4.NewCipher([]byte(key))
			if err != nil {
				log.Warn("tigerrat rc4 error: %v", err)
				h.delegate.CloseConnection()
				return
			}

			h.sendCipher, err = rc4.NewCipher([]byte(key))
			if err != nil {
				log.Warn("tigerrat rc4 error: %v", err)
				h.delegate.CloseConnection()
				return
			}

			h.delegate.SendData([]byte("xPPygOn\x00"))

			h.handshakeComplete = true
		} else {
			cmd := command{}

			b, err := queue.Get(h.dataChan, 4)
			if err != nil {
				h.delegate.CloseConnection()
				return
			}

			size := binary.LittleEndian.Uint32(b[0:4])
			if size > 0 {
				b, err = queue.Get(h.dataChan, int(size))
				if err != nil {
					h.delegate.CloseConnection()
					return
				}

				decrypted := make([]byte, size)
				h.recvCipher.XORKeyStream(decrypted, b)

				cmd.module = binary.LittleEndian.Uint32(decrypted[0:4])
				cmd.opcode = binary.LittleEndian.Uint32(decrypted[4:8])
				cmd.size = binary.LittleEndian.Uint32(decrypted[8:12])
				cmd.data = make([]byte, cmd.size)
				copy(cmd.data, decrypted[12:])
				h.proccessCommand(cmd)
			}
		}
	}
}

func (h *Handler) proccessCommand(cmd command) {
	logCommand(cmd)

	switch cmd.module {
	case modMain:
		switch cmd.opcode {
		case 0x1:
			hash := sha256.Sum256(cmd.data)
			id := hex.EncodeToString(hash[:])

			h.delegate.AgentConnected(id)
		}
	case modShell:
		switch cmd.opcode {
		case 0x11:
			log.Info(decodeWideString(cmd.data))
		case 0x12:
			log.Success("Execute complete")
		case 0x32:
			log.Warn("Execute failed")
		}
	case modFileManager:
		switch cmd.opcode {
		case 0x41:
			buf := make([]byte, 0x10000)
			for {
				bytesRead, err := h.file.Read(buf)

				if err != nil {
					if err != io.EOF {
						log.Warn("Error reading source file; %v", err)
					}

					break
				}

				h.sendCommand(command{
					module: modFileManager,
					opcode: opFileManagerUploadData,
					size:   uint32(bytesRead),
					data:   buf,
				})
			}

			// Finish the file transfer
			h.sendCommand(command{
				module: modFileManager,
				opcode: opFileManagerUploadDone,
				size:   0x0,
			})

			log.Success("Upload complete")
		case 0x44:
			if h.file != nil {
				h.file.Close()
				h.file = nil
				h.fileName = ""
			}
			log.Warn("Upload failed")
		case 0x51:
			file, err := os.Create(h.fileName)
			if err != nil {
				log.Warn("Error opening destination file: %v", err)
				return
			}
			h.file = file
		case 0x52:
			log.Warn("Download failed")
		case 0x53:
			if h.file != nil {
				h.file.Write(cmd.data)
			}
		case 0x54:
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
	log.Debug("TigerRAT Command")
	log.Debug("Module: 0x%x", c.module)
	log.Debug("Opcode: 0x%x", c.opcode)
	log.Debug("  Size: 0x%x", c.size)
	if c.size > 0 {
		log.Debug("  Data:\n%s", hex.Dump(c.data))
	}
}

func (h *Handler) sendCommand(cmd command) error {
	data := make([]byte, 12+len(cmd.data))

	binary.LittleEndian.PutUint32(data[0:4], cmd.module)
	binary.LittleEndian.PutUint32(data[4:8], cmd.opcode)
	binary.LittleEndian.PutUint32(data[8:12], cmd.size)

	if cmd.size > 0 {
		copy(data[12:], cmd.data)
	}

	packet := make([]byte, 16+len(cmd.data))
	binary.LittleEndian.PutUint32(packet[0:4], uint32(len(data)))
	encrypted := make([]byte, len(data))
	h.sendCipher.XORKeyStream(encrypted, data)
	copy(packet[4:], encrypted)

	h.delegate.SendData(packet)

	return nil
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
