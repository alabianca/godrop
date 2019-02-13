package godrop

import (
	"bufio"
	"compress/gzip"
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
	BUF_SIZE             = 1024
	PADDING              = "/"
)

// Session represents the connection between 2 peers
type Session struct {
	conn        net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	isEncrypted bool
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

// IsEncrypted returns true if the underlying tcp connection is encrypted
func (s *Session) IsEncrypted() bool {
	return s.isEncrypted
}

// Close closes the underlying tcp connection
func (s *Session) Close() {
	s.conn.Close()
}

// TransferContent writes the contents of dir into a gzip writer.
// The underlying gzip writer writes into s.writer which represents the underlying
// tcp connection
//
// tarWriter ---> gzipWriter ---> bufioWriter ---> net.Conn
func (s *Session) TransferContent(dir string) error {

	gzw := gzip.NewWriter(s.writer)

	defer s.Close()        // close underlying tcp conn
	defer s.writer.Flush() // close buffio.Writer
	defer gzw.Close()      // close gzip writer

	WriteTarball(gzw, dir)

	return nil
}

// CloneDir clones the advertised directory into dir
//
// tarReader <--- gzipReader <--- bufioReader <--- net.Conn
func (s *Session) CloneDir(dir string) error {

	gzr, err := gzip.NewReader(s.reader)

	if err != nil {
		return err
	}

	return ReadTarball(gzr, dir)
}
