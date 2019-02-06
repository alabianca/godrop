package godrop

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
)

const (
	HANDSHAKE_LENGTH     = 275
	HANDSHAKE_ACK_LENGTH = 6
	HANDSHAKE            = 0x3C
	HANDSHAKE_SYN_ACK    = 0x3D
	HANDSHAKE_ACK        = 0x3E
	END_OF_TEXT          = 0x3
)

// Session represents the connection between 2 peers
type Session struct {
	reader        *bufio.Reader
	writer        *bufio.Writer
	myPublicKey   *rsa.PublicKey
	myPrivateKey  *rsa.PrivateKey
	peerPublicKey *rsa.PublicKey
	isEncrypted   bool
}

// NewSession returns a new session instance.
// If private or public key pairs are nil, session will be unencrypted.
func NewSession(conn net.Conn, clientFlag bool, privateKey *rsa.PrivateKey, pubKey *rsa.PublicKey) (*Session, error) {
	sesh := new(Session)
	sesh.reader = bufio.NewReader(conn)
	sesh.writer = bufio.NewWriter(conn)

	if privateKey != nil && pubKey != nil {
		sesh.isEncrypted = true
		sesh.myPublicKey = pubKey
		sesh.myPrivateKey = privateKey

	} else {
		sesh.isEncrypted = false
	}

	if sesh.isEncrypted {
		if err := sesh.doHandshake(clientFlag); err != nil {
			return nil, err
		}

	}

	return sesh, nil
}

func (s *Session) doHandshake(clientFlag bool) error {
	if clientFlag {
		if err := s.Handshake(); err != nil {
			return err
		}
		return nil
	}

	if err := s.readHandshake(); err != nil {
		return err
	}
	return nil
}

func (s *Session) readHandshake() error {
	buf := make([]byte, HANDSHAKE_LENGTH)
	response := make([]byte, 0)
	totalRead := 0

	for {
		n, err := s.reader.Read(buf)

		if err != nil {
			return err
		}

		if totalRead < HANDSHAKE_LENGTH {
			response = append(response, buf[:n]...)
			totalRead += n
		}

		if totalRead >= HANDSHAKE_LENGTH {
			break
		}
	}

	msg := new(Message)
	msg.Decode(response)

	pubKey, err := x509.ParsePKCS1PublicKey(msg.payload)

	if err != nil {
		return err
	}

	s.peerPublicKey = pubKey

	if err := s.HandshakeSYNACK(); err != nil {
		return err
	}

	return nil
}

// Handshake initiates the 3-way handshake to exchange public keys with the peer
// It will return an error if no public/private keys are set on the session
// Handshake will send a Handshake packet to the peer and will block until the peer
// responds with a Handshake SYN ACK packet. The final packet is a Handshake ACK packet to the peer
func (s *Session) Handshake() error {
	if !s.isEncrypted {
		return fmt.Errorf("No Public/Private Key Pair set")
	}

	key := x509.MarshalPKCS1PublicKey(s.myPublicKey)
	msg := NewMessage(HANDSHAKE, key)

	packet := msg.Encode()

	n, err := s.writer.Write(packet)

	if err != nil || n < len(packet) {
		return err
	}

	s.writer.Flush()

	// wait for the response
	buf := make([]byte, HANDSHAKE_LENGTH)
	x, err := s.reader.Read(buf)

	if err != nil {
		return err
	}

	if x != HANDSHAKE_LENGTH {
		return fmt.Errorf("handshake error")
	}

	response := new(Message)
	response.Decode(buf)

	if response.header.Type != HANDSHAKE_SYN_ACK {
		return fmt.Errorf("Handshake failed")
	}

	// finally set the peer public key
	peerKey, err := x509.ParsePKCS1PublicKey(response.payload)

	if err != nil {
		return err
	}

	s.peerPublicKey = peerKey

	s.handshakeACK()

	return nil
}

// handshakeACK is the last message in the 3-way handshake
func (s *Session) handshakeACK() error {
	msg := NewMessage(HANDSHAKE_ACK, []byte("OK"))
	packet := msg.Encode()

	n, err := s.writer.Write(packet)

	if err != nil || n < len(packet) {
		return err
	}

	s.writer.Flush()

	return nil

}

// HandshakeSYNACK sends a Handshake syn ack packet to the peer.
// HandshakeSYNACK returns an error if no public/private key pai is set
func (s *Session) HandshakeSYNACK() error {
	if !s.isEncrypted {
		return fmt.Errorf("No Public/Private Key Pair set")
	}

	key := x509.MarshalPKCS1PublicKey(s.myPublicKey)
	msg := NewMessage(HANDSHAKE_SYN_ACK, key)

	packet := msg.Encode()

	n, err := s.writer.Write(packet)

	if err != nil || n < len(packet) {
		return err
	}

	s.writer.Flush()

	// wait for the response
	buf := make([]byte, HANDSHAKE_ACK_LENGTH)
	x, err := s.reader.Read(buf)

	if err != nil {
		return err
	}

	if x != HANDSHAKE_ACK_LENGTH {
		return fmt.Errorf("handshake error")
	}

	response := new(Message)
	response.Decode(buf)

	if response.header.Type != HANDSHAKE_ACK {
		return fmt.Errorf("Handshake failed")
	}

	return nil

}

func (s *Session) IsEncrypted() bool {
	return s.isEncrypted
}

func (s *Session) Write(p []byte) (n int, err error) {
	n, err = s.writer.Write(p)
	return
}

func (s *Session) Flush() error {
	return s.writer.Flush()
}

func (s *Session) Read(buf []byte) (n int, err error) {
	n, err = s.reader.Read(buf)

	return
}
