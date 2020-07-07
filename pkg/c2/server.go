package c2

import (
	"net"
	"sync"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

// A Server represents a running mock C2 server.
type Server struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	handler  ProtocolHandler
}

// NewServer creates a new mock C2 server and starts it listening on the
// provided address.
func NewServer(protocol string, address string) (*Server, error) {
	handler, err := NewProtocolHandler(protocol)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		quit:     make(chan interface{}),
		handler:  handler,
	}

	handler.SetDelegate(s)

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
			go func() {
				a, err := s.handler.ValidateConnection(conn, s.quit)
				if err != nil {
					log.Warn(err.Error())
				} else {
					AddAgent(a)
					s.handler.HandleConnection(conn, s.quit)
				}
				s.wg.Done()
			}()
		}
	}
}
