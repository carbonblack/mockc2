package c2

import (
	"fmt"
	"net"
	"sync"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

// A Server represents a running mock C2 server.
type Server struct {
	listener        net.Listener
	quit            chan interface{}
	wg              sync.WaitGroup
	port            uint16
	protocolHandler ProtocolHandler
}

// NewServer creates a new mock C2 server and starts it listening on the
// provided port.
func NewServer(port uint16, protocolHandler ProtocolHandler) (*Server, error) {
	s := &Server{
		quit:            make(chan interface{}),
		port:            port,
		protocolHandler: protocolHandler,
	}

	address := fmt.Sprintf(":%d", s.port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	s.listener = l
	log.Debug("Server listening")

	s.wg.Add(1)
	go s.serve()

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
			go func() {
				a, err := s.protocolHandler.ValidateConnection(conn, s.quit)
				if err != nil {
					log.Warn(err.Error())
				} else {
					AddAgent(a)
					s.protocolHandler.HandleConnection(conn, s.quit)
				}
				s.wg.Done()
			}()
		}
	}
}
