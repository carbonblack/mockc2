package c2

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"net"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/internal/queue"
)

const (
	hcBeacon = 0x7c8
)

// HotCroissant is a protocol handler capable of communicating with the
// HotCroissant malware family.
type HotCroissant struct {
	delegate ProtocolDelegate
	dataChan chan byte
}

type hcEncodedCommand struct {
	compressedSize   uint32
	uncompressedSize uint32
	data             []byte
}

type hcCommand struct {
	opcode uint32
	txnID  int32
	opt1   int32
	opt2   int32
	opt3   int32
	size   uint32
	data   []byte
}

// SetDelegate saves the delegate for later use.
func (h *HotCroissant) SetDelegate(delegate ProtocolDelegate) {
	h.delegate = delegate
}

// ReceiveData saves the network data and processes it when a full command has
// been received.
func (h *HotCroissant) ReceiveData(data []byte) {
	log.Debug("received\n" + hex.Dump(data))

	if h.dataChan == nil {
		h.dataChan = make(chan byte, len(data))
		go h.processData()
	}

	queue.Put(h.dataChan, data)
}

// Close cleans up any uzed resources.
func (h *HotCroissant) Close() {
	close(h.dataChan)
}

func (h *HotCroissant) processData() {
	for {
		ec := hcEncodedCommand{}

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

func (h *HotCroissant) proccessCommand(command hcCommand) {
	logCommand(command)

	switch command.opcode {
	case hcBeacon:
		a := &Agent{}
		hash := sha256.Sum256(command.data)
		a.ID = hex.EncodeToString(hash[:])

		h.delegate.AgentConnected(a)
	}
}

func logCommand(c hcCommand) {
	log.Debug("HotCroissant Command")
	log.Debug("Opcode: 0x%08x", c.opcode)
	log.Debug("  Opt1: 0x%08x", c.opt1)
	log.Debug("  Opt2: 0x%08x", c.opt2)
	log.Debug("  Opt3: 0x%08x", c.opt3)
	log.Debug("  Size: 0x%08x", c.size)
	log.Debug("  Data:\n%s", hex.Dump(c.data))
}

func decodeCommand(ec hcEncodedCommand) (hcCommand, error) {
	decrypted := cipher(ec.data)

	decompressed, err := decompress(decrypted)
	if err != nil {
		return hcCommand{}, err
	}

	buf := bytes.NewReader(decompressed)

	c := hcCommand{}
	err = binary.Read(buf, binary.LittleEndian, &c.opcode)
	if err != nil {
		return hcCommand{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt1)
	if err != nil {
		return hcCommand{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt2)
	if err != nil {
		return hcCommand{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.opt3)
	if err != nil {
		return hcCommand{}, err
	}

	err = binary.Read(buf, binary.LittleEndian, &c.size)
	if err != nil {
		return hcCommand{}, err
	}

	c.data = decompressed[20:]

	return c, nil
}

func sendCommand(conn net.Conn, c hcCommand) error {
	ec, err := encodeCommand(c)
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.LittleEndian, &ec.compressedSize)
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.LittleEndian, &ec.uncompressedSize)
	if err != nil {
		return err
	}

	_, err = conn.Write(ec.data)
	if err != nil {
		return err
	}

	return nil
}

func encodeCommand(c hcCommand) (hcEncodedCommand, error) {
	var b bytes.Buffer

	binary.Write(&b, binary.LittleEndian, c.opcode)
	binary.Write(&b, binary.LittleEndian, c.opt1)
	binary.Write(&b, binary.LittleEndian, c.opt2)
	binary.Write(&b, binary.LittleEndian, c.opt3)
	binary.Write(&b, binary.LittleEndian, c.size)

	data := append(b.Bytes(), c.data...)

	log.Debug("encoded\n" + hex.Dump(data))

	ec := hcEncodedCommand{}
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
