package c2

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"time"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

// A Generic protocol handler simply logs information about connections and
// the data received.
type Generic struct {
}

// ValidateConnection makes sure a real agent is connecting and returns a
// new instance of the Agent. In the case of the Generic protocol, all
// connections are accepted.
func (g Generic) ValidateConnection(conn net.Conn, quit chan interface{}) (*Agent, error) {
	addr := conn.RemoteAddr().String()
	h := sha256.Sum256([]byte(addr))
	id := hex.EncodeToString(h[:])

	a := &Agent{
		ID:   id,
		Addr: conn.RemoteAddr(),
	}

	return a, nil
}

// HandleConnection handles generic connections by logging.
func (g Generic) HandleConnection(conn net.Conn, quit chan interface{}) {
	defer conn.Close()

	log.Info("connection from %v", conn.RemoteAddr())

	buf := make([]byte, 2048)

	for {
		select {
		case <-quit:
			return
		default:
			conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
			n, err := conn.Read(buf)
			if n > 0 {
				log.Debug("received\n" + hex.Dump(buf[:n]))
			}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				} else if err != io.EOF {
					log.Warn("read error %v", err)
					return
				}
			}
			if n == 0 {
				return
			}
		}
	}
}
