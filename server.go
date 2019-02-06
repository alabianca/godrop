package godrop

import (
	"crypto/rsa"
	"fmt"
	"net"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	IP          string
	mdnsService *zeroconf.Server
	listener    net.Listener
	shutdown    chan struct{}
	pubKey      *rsa.PublicKey
	prvKey      *rsa.PrivateKey
}

func (s *Server) Shutdown() {
	fmt.Println("Shutting down...")
	s.mdnsService.Shutdown()
	close(s.shutdown)
}

func (s *Server) listen() error {
	port := strconv.Itoa(s.Port)
	address := net.JoinHostPort(s.IP, port)

	// todo handle a listen error properly
	l, err := net.Listen("tcp4", address)

	s.listener = l

	return err
}

func (s *Server) Accept() (*Session, error) {
	conn, err := s.listener.Accept()

	if err != nil {
		return nil, err
	}

	sesh, err := NewSession(conn, false, s.prvKey, s.pubKey)

	return sesh, err
}
