package godrop

import (
	"net"
)

type P2PConn struct {
	conn *net.TCPConn
}

func (c *P2PConn) Read(b []byte) (n int, err error) {
	n, err = c.conn.Read(b)

	return
}

func (c *P2PConn) Write(b []byte) (n int, err error) {
	n, err = c.conn.Write(b)

	return
}

func (c *P2PConn) Close() {
	c.conn.Close()
}
