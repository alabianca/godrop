package godrop

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	HANDSHAKE_LENGTH     = 275
	HANDSHAKE_ACK_LENGTH = 6
	HANDSHAKE            = 0x3C
	HANDSHAKE_SYN_ACK    = 0x3D
	HANDSHAKE_ACK        = 0x3E
	END_OF_TEXT          = 0x3
	MESSAGE              = 0x3F
	BUF_SIZE             = 1024
	PADDING              = "/"
)

// Session represents the connection between 2 peers
type Session struct {
	conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	Finfo         os.FileInfo
	isEncrypted   bool
	sessionHeader Header
}

// NewSession returns a new session instance.
// If private or public key pairs are nil, session will be unencrypted.
func NewSession(conn net.Conn, clientFlag bool) (*Session, error) {
	sesh := new(Session)
	sesh.conn = conn
	sesh.reader = bufio.NewReader(conn)
	sesh.writer = bufio.NewWriter(conn)
	sesh.isEncrypted = false

	return sesh, nil
}

func (s *Session) IsEncrypted() bool {
	return s.isEncrypted
}

func (s *Session) Close() {
	s.conn.Close()
}

// WriteHeader writes the header packet to the peer
// A  header contains the file size and file name
func (s *Session) WriteHeader() {
	bufSize := fillString(strconv.FormatInt(s.Finfo.Size(), 10), 10)
	bufFName := fillString(s.Finfo.Name(), 64)

	s.writer.Write([]byte(bufSize))
	s.writer.Write([]byte(bufFName))
	s.writer.Flush()

}

func (s *Session) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

// Flush writes any buffered data in the underlying io.Writer
func (s *Session) Flush() error {
	return s.writer.Flush()
}

func (s *Session) Read(buf []byte) (n int, err error) {
	return s.reader.Read(buf)
}

// ReadHeader reads the header packet from the session
// A header contains the file name and the file size that is being transferred
func (s *Session) ReadHeader() (Header, error) {
	contentLength := make([]byte, 10)

	if _, err := s.reader.Read(contentLength); err != nil {
		return Header{}, err
	}

	contentName := make([]byte, 64)

	if _, err := s.reader.Read(contentName); err != nil {
		return Header{}, err
	}

	fileSize, err := strconv.ParseInt(strings.Trim(string(contentLength), PADDING), 10, 64)

	if err != nil {
		return Header{}, err
	}

	header := Header{
		Size: fileSize,
		Name: strings.Trim(string(contentName), PADDING),
	}

	s.sessionHeader = header

	return header, nil

}
