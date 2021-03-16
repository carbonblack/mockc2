package yort

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

// TODO since HTTP is stateless we really need command queuing in order to send
// commands to the agent when it connects and says it's ready for commands

// Handler is a Yort protocol handler.
type Handler struct {
	delegate        protocol.Delegate
	pr              *io.PipeReader
	pw              *io.PipeWriter
	b               *bufio.Reader
	downloadStarted bool
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

	if h.b == nil {
		h.pr, h.pw = io.Pipe()
		h.b = bufio.NewReader(h.pr)
		go h.processData()
	}

	h.pw.Write(data)
}

func (h *Handler) processData() {
	for {
		req, err := http.ReadRequest(h.b)
		if err != nil && err != io.EOF {
			log.Warn("Yort error reading request: %v", err)
		}

		if err == io.EOF {
			break
		}

		if req != nil {
			sessid := req.FormValue("_webident_f")
			value := req.FormValue("_webident_s")

			log.Debug("Session ID: %s Value: %s\n", sessid, value)

			if value != "" {
				switch value {
				case "16":
					log.Debug("Session created")

					h.sendHTTPResponse([]byte("1"))
				case "17":
					log.Debug("Session destroyed")

					h.sendHTTPResponse([]byte("1"))
				case "20", "21":
					log.Debug("Data received")

					var buf bytes.Buffer
					file, header, err := req.FormFile("file")
					if err != nil {
						log.Debug("No file form value", err)
					} else {
						log.Debug("File name %s\n", header.Filename)

						io.Copy(&buf, file)
						decrypted := cipher(buf.Bytes())
						fmt.Println(hex.Dump(decrypted))
						file.Close()

						if value == "21" {
							log.Debug("Skipping response")
						} else {
							if decrypted[0] == 0x05 {
								if h.downloadStarted {
									// Only if we started a download we need to send offset
									h.sendCommand(0x5, []byte{0x00, 0x00, 0x00, 0x00})
									h.downloadStarted = false
								} else {
									h.sendCommand(0x5, nil)
								}
							} else {
								h.sendCommand(0x5, nil)
							}
						}
					}
				case "22":
					log.Debug("Ready for commands")
					h.sendCommand(0xb, nil) // system info
					// writeCommand(res, 0x2, []byte{0x01, 0x00, 0x00, 0x00}) // sleep (in munutes)
					// writeCommand(res, 0x3, nil) // die
					// writeCommand(res, 0xb, nil) // system info
					// writeCommand(res, 0xc, nil) // keep alive
					// writeCommand(res, 0xe, nil) // get config length is 0x868
					// writeCommand(res, 0xf, nil) // set config length is 0x868
					// writeCommand(res, 0x12, []byte("ls -la")) // shell bin/bash
					// writeCommand(res, 0x13, []byte("whoami")) // shell using fork
					// writeCommand(res, 0x15, []byte("/Users/fakeuser/Desktop/test.txt"))
				}
			}
		}
	}
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	log.Warn("yort doesn't support command execution")
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	log.Warn("yort doesn't support file upload")
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	log.Warn("yort doesn't support file download")
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
	h.pw.Close()
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return true
}

func (h *Handler) sendHTTPResponse(data []byte) {
	resp := http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBuffer(data)),
		ContentLength: int64(len(data)),
		Header:        make(http.Header, 0),
	}

	buff := bytes.NewBuffer(nil)
	resp.Write(buff)

	h.delegate.SendData(buff.Bytes())
}

func (h *Handler) sendCommand(opcode uint32, data []byte) {
	command := make([]byte, 12+len(data))
	binary.LittleEndian.PutUint32(command[0:4], opcode)
	binary.LittleEndian.PutUint32(command[4:8], 0x0)
	binary.LittleEndian.PutUint32(command[8:12], uint32(len(data)))
	if len(data) > 0 {
		copy(command[12:], data)
	}

	encrypted := cipher(command)

	h.sendHTTPResponse(encrypted)
}

func cipher(input []byte) []byte {
	output := make([]byte, len(input))

	for i := 0; i < len(input); i++ {
		output[i] = input[i] ^ 0xaa
	}

	return output
}
