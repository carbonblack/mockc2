package c2

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/bistromath"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/generic"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/hotcroissant"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/mata"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/rifdoor"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol/slickshoes"
)

// A Server represents a running mock C2 server.
type Server struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	protocol string
}

type c2Conn struct {
	conn    net.Conn
	quit    chan interface{}
	handler protocol.Handler
}

func handlerFromString(protocol string) (protocol.Handler, error) {
	switch protocol {
	case "bistromath":
		return &bistromath.Handler{}, nil
	case "generic":
		return &generic.Handler{}, nil
	case "hotcroissant":
		return &hotcroissant.Handler{}, nil
	case "mata":
		return &mata.Handler{}, nil
	case "rifdoor":
		return &rifdoor.Handler{}, nil
	case "slickshoes":
		return &slickshoes.Handler{}, nil
	default:
		return nil, fmt.Errorf("unknown protocol %s", protocol)
	}
}

// NewServer creates a new mock C2 server and starts it listening on the
// provided address.
func NewServer(protocol string, address string) (*Server, error) {
	handler, err := handlerFromString(protocol)
	if err != nil {
		return nil, err
	}

	s := &Server{
		quit:     make(chan interface{}),
		protocol: protocol,
	}

	var l net.Listener
	if handler.NeedsTLS() {
		certPEM, keyPEM, err := createCertificate()
		if err != nil {
			return nil, err
		}

		cer, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, err
		}

		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		l, err = tls.Listen("tcp", address, config)
	} else {
		l, err = net.Listen("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	s.listener = l
	s.wg.Add(1)
	go s.serve()

	log.Debug("Server listening")

	return s, nil
}

// Shutdown gracefully shuts down the C2 server without interrupting any
// active connections.
func (s *Server) Shutdown() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Warn("accept error %v", err)
			}
		} else {
			s.wg.Add(1)

			c, err := newC2Conn(conn, s)
			if err != nil {
				return
			}

			go func() {
				defer s.wg.Done()
				c.receiveLoop()
			}()
		}
	}
}

func newC2Conn(netConn net.Conn, s *Server) (*c2Conn, error) {
	handler, err := handlerFromString(s.protocol)
	if err != nil {
		return nil, err
	}

	c := &c2Conn{
		conn:    netConn,
		quit:    s.quit,
		handler: handler,
	}

	c.handler.SetDelegate(c)

	return c, nil
}

func (c *c2Conn) receiveLoop() {
	defer c.conn.Close()

	log.Info("connection from %v", c.conn.RemoteAddr())
	c.handler.Accept()

	buf := make([]byte, 2048)

	for {
		select {
		case <-c.quit:
			c.handler.Close()
			return
		default:
			c.conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
			n, err := c.conn.Read(buf)
			if n > 0 {
				c.handler.ReceiveData(buf[:n])
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

func (c *c2Conn) SendData(data []byte) {
	_, err := c.conn.Write(data)
	if err != nil {
		log.Warn("write error %v", err)
	}

	// TODO handle writing less than the total bytes of data

	log.Debug("sent\n" + hex.Dump(data))
}

func (c *c2Conn) CloseConnection() {
	c.conn.Close()
}

func (c *c2Conn) AgentConnected(id string) {
	a := &Agent{
		ID: id,
	}

	// Default Agent ID to a hash of the IP if not specified
	if a.ID == "" {
		addr := c.conn.RemoteAddr().String()
		h := sha256.Sum256([]byte(addr))
		a.ID = hex.EncodeToString(h[:])
	}

	a.Addr = c.conn.RemoteAddr()
	a.conn = c

	AddAgent(a)
}
