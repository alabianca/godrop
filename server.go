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
	sharePath   string
	fInfo       os.FileInfo
	shutdown    chan struct{}
}

func (s *Server) Shutdown() {
	fmt.Println("Shutting down...")
	s.mdnsService.Shutdown()
	close(s.shutdown)
}

func (s *Server) Start() {
	listener, err := s.listen()

	if err != nil {
		s.Shutdown()
		return
	}

	for {
		sesh, err := s.accept(listener)

		if err != nil {
			continue
		}

		go s.handleConnection(sesh)
	}
}

func (s *Server) listen() (net.Listener, error) {
	port := strconv.Itoa(s.Port)
	address := net.JoinHostPort(s.IP, port)

	// todo handle a listen error properly
	l, err := net.Listen("tcp4", address)

	return l, err
}

func (s *Server) accept(l net.Listener) (*Session, error) {
	conn, err := l.Accept()

	if err != nil {
		return nil, err
	}

	sesh, err := NewSession(conn, false)

	return sesh, err
}

func (s *Server) handleConnection(session *Session) {
	if err := transferDir(session, s.sharePath); err != nil {
		fmt.Println(err)
	}
}

func transferDir(session *Session, dir string) error {

	return session.TransferContent(dir)
	//return WriteTarball(session.writer, dir)
}
