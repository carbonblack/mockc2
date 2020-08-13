package crosswalk

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type uuid [16]byte

func newUUID() (uuid, error) {
	var u uuid
	_, err := io.ReadFull(rand.Reader, u[:])
	if err != nil {
		return u, err
	}
	u[6] = (u[6] & 0x0f) | 0x40 // Version 4
	u[8] = (u[8] & 0x3f) | 0x80 // Variant is 10
	return u, nil
}

func (u uuid) String() string {
	// This is not the standard way to convert a UUID to a String
	// This is however how Crosswalk does it.
	// Crosswalk does it wrong and ends up printing the stack address
	// for data4 into the UUID
	data1 := binary.LittleEndian.Uint32(u[0:4])
	data2 := binary.LittleEndian.Uint16(u[4:6])
	data3 := binary.LittleEndian.Uint16(u[6:8])
	data4 := binary.LittleEndian.Uint64(u[8:])

	return fmt.Sprintf("%08X-%04X-%04X-%011X", data1, data2, data3, data4)
}

func (u uuid) Hash() [16]byte {
	s := []byte(u.String())
	length := 72 - len(s)
	padded := append(s, bytes.Repeat([]byte{0x00}, length)...)

	return md5.Sum(padded)
}

func parseUUID(input []byte) (string, error) {
	if len(input) != 36 {
		return "", errors.New("uuid length is not 36 bytes")
	}

	if input[8] == 0x2d && input[13] == 0x2d && input[18] == 0x2d {
		return parse32bitUUID(input)
	}

	return parse64bitUUID(input)
}

func parse32bitUUID(input []byte) (string, error) {
	return string(input[:]), nil
}

func parse64bitUUID(input []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(string(input[:32]))
	if err != nil {
		return "", err
	}

	var u uuid
	copy(u[:], data[:16])

	return u.String(), nil
}
