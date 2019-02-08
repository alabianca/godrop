package godrop

import (
	"bufio"
	"net"
)

const (
	HANDSHAKE_LENGTH     = 275
	HANDSHAKE_ACK_LENGTH = 6
	HANDSHAKE            = 0x3C
	HANDSHAKE_SYN_ACK    = 0x3D
	HANDSHAKE_ACK        = 0x3E
	END_OF_TEXT          = 0x3
	MESSAGE              = 0x3F
)

// Session represents the connection between 2 peers
type Session struct {
	reader      *bufio.Reader
	writer      *bufio.Writer
	isEncrypted bool
}

// NewSession returns a new session instance.
// If private or public key pairs are nil, session will be unencrypted.
func NewSession(conn net.Conn, clientFlag bool) (*Session, error) {
	sesh := new(Session)
	sesh.reader = bufio.NewReader(conn)
	sesh.writer = bufio.NewWriter(conn)
	sesh.isEncrypted = false

	return sesh, nil
}

func (s *Session) IsEncrypted() bool {
	return s.isEncrypted
}

func (s *Session) Write(p []byte) (n int, err error) {
	if s.isEncrypted {
		return s.writeEncrypted(p)
	}

	return s.writeUnencrypted(p)
}

func (s *Session) Flush() error {
	return s.writer.Flush()
}

func (s *Session) writeEncrypted(p []byte) (n int, err error) {

	return 0, nil
}

func (s *Session) writeUnencrypted(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

func (s *Session) Read(buf []byte) (n int, err error) {
	n, err = s.reader.Read(buf)

	return
}
