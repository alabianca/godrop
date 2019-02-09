package godrop

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	IP          string
	mdnsService *zeroconf.Server
	listener    net.Listener
	sharePath   string
	shutdown    chan struct{}
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

func (s *Server) ReadInSharePath() (*os.File, error) {
	return os.Open(s.sharePath)
}

func (s *Server) Accept() (*Session, error) {
	conn, err := s.listener.Accept()

	if err != nil {
		return nil, err
	}

	sesh, err := NewSession(conn, false)

	return sesh, err
}
