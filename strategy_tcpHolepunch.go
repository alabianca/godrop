package godrop

import (
	"net"

	"github.com/alabianca/holepunch"
)

type TcpHolepunch struct {
	RelayIP      string
	RelayPort    string
	ListenAddr   string
	LocalPort    string
	LocalRelayIP string
	UID          string
}

func (h TcpHolepunch) Connect(peer string) (*P2PConn, error) {

	conf := holepunch.Config{
		RelayIP:      h.RelayIP,
		RelayPort:    h.RelayPort,
		ListenAddr:   h.ListenAddr,
		LocalPort:    h.LocalPort,
		LocalRelayIP: h.LocalRelayIP,
		UID:          h.UID,
	}

	punch, err := holepunch.NewHolepunch(conf)

	if err != nil {
		return nil, err
	}

	conn, err := punch.Connect(peer)

	if err != nil {
		return nil, err
	}

	p2pConn := &P2PConn{
		conn: conn.(*net.TCPConn),
	}

	return p2pConn, nil
}
