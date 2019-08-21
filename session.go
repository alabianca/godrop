package godrop

import (
	"bufio"
	"compress/gzip"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	AUTH_PACKET          = 60
	CLONE_PACKET         = 61
	CLONE_PACKET_ACK     = 62
	CLONE_PACKET_NAK     = 63
	AUTH_PACKET_LENGTH   = 3
	CLONE_ACKNACK_LENGTH = 65
	KEY_PACKET           = 66
	BUF_SIZE             = 1024
	PADDING              = "/"
)

// Session represents the connection between 2 peers
type Session struct {
	conn            net.Conn
	reader          *bufio.Reader
	writer          *bufio.Writer
	DebugWriter     io.Writer
	LocalHost       string
	RemoteHost      string
	RemoteService   string
	RemotePort      int
	RemoteIP        string
	RemoteDroplet   string
	isAuthenticated bool
	isEncrypted     bool
}

// NewSession returns a new session instance.
// If private or public key pairs are nil, session will be unencrypted.
func NewSession(conn net.Conn, clientFlag, isEncrypted bool) (*Session, error) {
	sesh := new(Session)
	sesh.conn = conn
	sesh.reader = bufio.NewReader(conn)
	sesh.writer = bufio.NewWriter(conn)
	sesh.isEncrypted = isEncrypted
	sesh.RemoteIP = conn.RemoteAddr().String()
	sesh.isAuthenticated = false

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

// Authenticate sends an AUTH_PACKET to the remote peer
//
// [1-Byte AUTH_PACKET][64 BYTES LOCAL HOST NAME]
func (s *Session) Authenticate() (Header, error) {
	authType := []byte{AUTH_PACKET}
	host := fillString(s.LocalHost, 64)

	s.writer.Write(authType)
	s.writer.Write([]byte(host))
	s.writer.Flush()

	response := make([]byte, 74)

	if _, err := io.ReadFull(s.reader, response); err != nil {
		return Header{}, err
	}

	size := response[:10]
	name := response[10:]
	contentLength, err := strconv.ParseInt(strings.Trim(string(size), PADDING), 10, 64)

	if err != nil {
		return Header{}, err
	}

	contentName := strings.Trim(string(name), PADDING)

	header := Header{
		Size: contentLength,
		Name: contentName,
	}

	return header, nil
}

func (s *Session) AuthenticateWithKey(key *rsa.PublicKey) (Header, error) {
	authType := []byte{KEY_PACKET}
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(len(pemBytes)))

	s.writer.Write(authType)
	s.writer.Write(b)
	s.writer.Write(pemBytes)
	s.writer.Flush()

	return Header{}, fmt.Errorf("Not Implemented")
}

// SendHeader sends a header packet to the remote peer
//
// [10-byte content-length][64-byte content name]
func (s *Session) SendHeader(name string, size int64) {
	contentLength := fillString(strconv.FormatInt(size, 10), 10)
	contentName := fillString(name, 64)

	s.writer.Write([]byte(contentLength))
	s.writer.Write([]byte(contentName))
	s.writer.Flush()
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
	// first send a packet to the peer indicating that we want to clone the content
	clonePacket := []byte{CLONE_PACKET}
	padding := fillString(PADDING, 64)
	s.writer.Write(clonePacket)
	s.writer.Write([]byte(padding))
	s.writer.Flush()

	// wait for the ack or nak
	response := make([]byte, CLONE_ACKNACK_LENGTH)
	if _, err := io.ReadFull(s.reader, response); err != nil {
		return err
	}

	// ACCESS-DENIED
	if response[0] == CLONE_PACKET_NAK {
		reason := response[1:]
		return fmt.Errorf(strings.Trim(string(reason), PADDING))
	}

	gzr, err := gzip.NewReader(s.reader)

	if err != nil {
		return err
	}

	if s.DebugWriter != nil {
		tr := io.TeeReader(gzr, s.DebugWriter)
		return ReadTarball(tr, dir)
	}

	return ReadTarball(gzr, dir)
}
