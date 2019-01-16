package godrop

import (
	"net"
)

type P2PConn struct {
	Conn *net.TCPConn
}

func (c *P2PConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)

	return
}

func (c *P2PConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)

	return
}

func (c *P2PConn) Close() {
	c.Conn.Close()
}
