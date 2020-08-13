package crosswalk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

func aesEncrypt(src []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Warn("Crosswalk key error1", err)
	}
	if len(src) == 0 {
		log.Warn("Crosswalk plain content empty")
	}
	ecb := cipher.NewCBCEncrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	content := src
	content = pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	return crypted
}

func aesDecrypt(crypt []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Warn("Crosswalk key error1", err)
	}
	if len(crypt) == 0 {
		log.Warn("Crosswalk plain content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return pkcs5Trimming(decrypted)
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func cryptDeriveKey(input []byte) []byte {
	hash := md5.Sum(input)

	var b0 []byte
	for _, b := range hash {
		b0 = append(b0, b^0x36)
	}

	var b1 []byte
	for _, b := range hash {
		b1 = append(b1, b^0x5c)
	}

	b0 = append(b0, bytes.Repeat([]byte{0x36}, 48)...)
	b1 = append(b1, bytes.Repeat([]byte{0x5c}, 48)...)

	b0md5 := md5.Sum(b0)
	b1md5 := md5.Sum(b1)

	finalKey := append(b0md5[:], b1md5[:]...)
	return finalKey[:16]
}

func generateKey(hash [16]byte) [16]byte {
	// Pad hash to 72 bytes
	length := 72 - len(hash)
	newInput := append(hash[:], bytes.Repeat([]byte{0x00}, length)...)
	derivedKey := cryptDeriveKey(newInput)

	encrypted := aesEncrypt(newInput, derivedKey)

	// 144 bytes total
	length = 144 - len(encrypted)
	padded := append(encrypted, bytes.Repeat([]byte{0x00}, length)...)
	finalKey := cryptDeriveKey(padded)

	var key [16]byte
	copy(key[:], finalKey)
	return key
}
