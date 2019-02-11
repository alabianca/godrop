package godrop

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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
	sesh.Finfo = s.fInfo

	return sesh, err
}

func (s *Server) handleConnection(session *Session) {
	transferDir(session, s.sharePath)

	// write one last packet signaling the end of the transfer before closing the connection
	session.WriteHeader(Header{
		Name:  "",
		Size:  0,
		Path:  "",
		Flags: isDoneMask,
	})

	session.Close()
}

func transferDir(session *Session, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header := Header{Name: info.Name(), Size: info.Size(), Path: path}

		if info.IsDir() {
			header.SetDirBit()
			session.WriteHeader(header)
			return nil // nothing more to write
		}

		session.WriteHeader(header)
		// content is a file. write the file now byte by byte
		file, err := os.Open(path)
		defer file.Close()

		if err != nil {
			return err
		}

		buf := make([]byte, BUF_SIZE)

		for {
			n, err := file.Read(buf)

			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

			session.Write(buf[:n])
			session.Flush()
		}

		return nil

	})
}
