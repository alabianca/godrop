package godrop

import (
	"fmt"
	"net"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	IP          string
	mdnsService *zeroconf.Server
	shutdown    chan struct{}
}

func (s *Server) Shutdown() {
	fmt.Println("Shutting down...")
	close(s.shutdown)
}

func (s *Server) Listen() (net.Listener, error) {
	port := strconv.Itoa(s.Port)
	address := net.JoinHostPort(s.IP, port)

	l, err := net.Listen("tcp4", address)

	return l, err

}

func (s *Server) handleConnection(conn *net.Conn) {
	fmt.Println("Got a connection ", conn)
}
