package c2

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/internal/queue"
)

const (
	rifdoorBeacon   uint32 = 0x9e2
	rifdoorRequest  uint32 = 0x4e3a
	rifdoorResponse uint32 = 0xa021
	rifdoorEnd      uint32 = 0x1055
)

// Rifdoor is a protocol handler capable of communicating with the Rifdoor
// malware family.
type Rifdoor struct {
	delegate ProtocolDelegate
	dataChan chan byte
	checksum uint32
}

type rifdoorCommand struct {
	opcode   uint32
	checksum uint32
	zero     uint32
	size     uint32
	data     []byte
}

// SetDelegate saves the delegate for later use.
func (r *Rifdoor) SetDelegate(delegate ProtocolDelegate) {
	r.delegate = delegate
}

// ReceiveData saves the network data and processes it when a full command has
// been received.
func (r *Rifdoor) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	if r.dataChan == nil {
		r.dataChan = make(chan byte, len(data))
		go r.processData()
	}

	queue.Put(r.dataChan, data)
}

// SendCommand sends a command to the connected agent.
func (r *Rifdoor) SendCommand(command interface{}) {
	switch c := command.(type) {
	case ExecuteCommand:
		commandLine := strings.TrimSpace(c.Name + " " + strings.Join(c.Args, " "))

		rc := rifdoorCommand{
			opcode:   rifdoorRequest,
			checksum: r.checksum,
			zero:     0x0,
			size:     uint32(len(commandLine)),
			data:     []byte(commandLine),
		}
		data := r.encodeCommand(rc)
		r.delegate.SendData(data)

		rc = rifdoorCommand{
			opcode:   rifdoorEnd,
			checksum: r.checksum,
			zero:     0x0,
			size:     0x0,
		}
		data = r.encodeCommand(rc)
		r.delegate.SendData(data)
	case UploadCommand:
		log.Warn("rifdoor doesn't support file upload")
	case DownloadCommand:
		log.Warn("rifdoor doesn't support file download")
	}
}

func (r *Rifdoor) encodeCommand(command rifdoorCommand) []byte {
	result := make([]byte, 16+command.size)

	binary.LittleEndian.PutUint32(result[0:], command.opcode)
	binary.LittleEndian.PutUint32(result[4:], command.checksum)
	binary.LittleEndian.PutUint32(result[8:], command.zero)
	binary.LittleEndian.PutUint32(result[12:], command.size)

	if command.size > 0 {
		copy(result[16:], rifdoorCipher(command.data))
	}

	return result
}

// Close cleans up any uzed resources.
func (r *Rifdoor) Close() {
	close(r.dataChan)
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (r *Rifdoor) NeedsTLS() bool {
	return false
}

func (r *Rifdoor) processData() {
	for {
		command := rifdoorCommand{}

		// Receive the header
		b, err := queue.Get(r.dataChan, 16)
		if err != nil {
			r.delegate.CloseConnection()
			return
		}

		command.opcode = binary.LittleEndian.Uint32(b[0:4])
		command.checksum = binary.LittleEndian.Uint32(b[4:8])
		command.zero = binary.LittleEndian.Uint32(b[8:12])
		command.size = binary.LittleEndian.Uint32(b[12:16])

		if command.size > 0 {
			b, err = queue.Get(r.dataChan, int(command.size))
			if err != nil {
				r.delegate.CloseConnection()
				return
			}

			command.data = rifdoorCipher(b)
		}

		r.processCommand(command)
	}
}

func (r *Rifdoor) processCommand(command rifdoorCommand) {
	r.logCommand(command)

	r.checksum = command.checksum

	switch command.opcode {
	case rifdoorBeacon:
		a := &Agent{}
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, command.checksum)
		hash := sha256.Sum256(data)
		a.ID = hex.EncodeToString(hash[:])

		r.delegate.AgentConnected(a)
	case rifdoorResponse:
		log.Info(string(command.data))
	case rifdoorEnd:
		r.delegate.CloseConnection()
	}
}

func (r *Rifdoor) logCommand(command rifdoorCommand) {
	log.Debug("Rifdoor Command")
	log.Debug("  Opcode: 0x%08x", command.opcode)
	log.Debug("Checksum: 0x%08x", command.checksum)
	log.Debug("    Zero: 0x%08x", command.zero)
	log.Debug("    Size: 0x%08x", command.size)
	if command.data != nil {
		log.Debug("    Data:\n%s", hex.Dump(command.data))
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

func rifdoorCipher(input []byte) []byte {
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
