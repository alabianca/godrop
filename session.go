package godrop

import (
	"bufio"
	"fmt"
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

// IsEncrypted returns true if the underlying tcp connection is encrypted
func (s *Session) IsEncrypted() bool {
	return s.isEncrypted
}

// Close closes the underlying tcp connection
func (s *Session) Close() {
	s.conn.Close()
}

// WriteHeader writes the header packet to the peer
// A  header contains the file size and file name
func (s *Session) WriteHeader(h Header) {

	bufSize := fillString(strconv.FormatInt(h.Size, 10), 10)
	bufFName := fillString(h.Name, 64)
	pathLength := fillString(strconv.FormatInt(int64(len(h.Path)), 10), 10)
	path := fillString(h.Path, len(h.Path))
	//fmt.Println(bufSize, bufFName)
	flags := []byte{byte(h.Flags)}

	s.writer.Write([]byte(bufSize))
	n, _ := s.writer.Write([]byte(bufFName))

	fmt.Printf("written: %d. expected: %d. t: %s\n", n, len(bufFName), bufFName)

	s.writer.Write(flags)
	s.writer.Write([]byte(pathLength))
	s.writer.Write([]byte(path))
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
	//fmt.Println("CL: ", string(contentLength))
	contentName := make([]byte, 64)

	if _, err := s.reader.Read(contentName); err != nil {
		return Header{}, err
	}
	//fmt.Println("Content: ", string(contentName))
	fileSize, err := strconv.ParseInt(strings.Trim(string(contentLength), PADDING), 10, 64)

	if err != nil {
		return Header{}, err
	}

	flags := make([]byte, 1)

	if _, err := s.reader.Read(flags); err != nil {
		return Header{}, err
	}
	//fmt.Println("Flags: ", string(flags))
	flagByte := int(flags[0])

	header := Header{
		Size: fileSize,
		Name: strings.Trim(string(contentName), PADDING),
	}

	if (flagByte & isDirMask) > 0 {
		header.SetDirBit()
	}
	if (flagByte & isDoneMask) > 0 {
		header.SetDoneBit()
	}

	pathLength := make([]byte, 10)

	if _, err := s.reader.Read(pathLength); err != nil {
		return Header{}, err
	}
	//fmt.Println("PathLength: ", string(pathLength))
	pathSize, err := strconv.ParseInt(strings.Trim(string(pathLength), PADDING), 10, 64)

	if err != nil {
		//fmt.Println(string(pathLength))
		return Header{}, err
	}

	path := make([]byte, pathSize)

	if _, err := s.reader.Read(path); err != nil {
		return Header{}, err
	}

	header.Path = strings.Trim(string(path), PADDING)

	s.sessionHeader = header

	return header, nil

}

func (s *Session) NewCloner() *Clone {
	c := &Clone{
		sesh:             s,
		readHeader:       make(chan int, 1),
		readContent:      make(chan Header, 1),
		transferComplete: make(chan int, 1),
	}

	c.readHeader <- 1 //immediately start reading the header

	return c
}
