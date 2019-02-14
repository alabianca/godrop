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

	return sesh, err
}

func (s *Server) handleConnection(session *Session) {
	// if err := transferDir(session, s.sharePath); err != nil {
	// 	fmt.Println(err)
	// }

	buf := make([]byte, AUTH_PACKET_LENGTH)

	for {
		_, err := io.ReadFull(session.reader, buf)

		if err != nil {
			return
		}

		msgType := buf[0]
		fmt.Println("got message: ", msgType)
		switch msgType {
		case AUTH_PACKET:
			//send header
			size, err := computeContentLength(s.sharePath)

			if err != nil {
				return
			}
			session.isAuthenticated = true
			session.SendHeader(s.sharePath, size)

		case CLONE_PACKET:
			// check if the session if authenticated
			if !session.isAuthenticated {
				nak := []byte{CLONE_PACKET_NAK}
				reason := fillString("Not Authenticated", 64)
				session.writer.Write(nak)
				session.writer.Write([]byte(reason))
				session.writer.Flush()
				continue
			}

			// session is authenticated. send ack and then transfer the contents
			ack := []byte{CLONE_PACKET_ACK}
			reason := fillString(PADDING, 64) //just an empty padding for now. as the client expects exactly 65 bytes
			session.writer.Write(ack)
			session.writer.Write([]byte(reason))

			if err := transferDir(session, s.sharePath); err != nil {
				return
			}

		}

	}
}

func computeContentLength(dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}

func transferDir(session *Session, dir string) error {

	return session.TransferContent(dir)
	//return WriteTarball(session.writer, dir)
}
