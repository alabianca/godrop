package godrop

import (
	"strings"
)

type Godrop struct {
	tcpServer     *server
	peer          *Peer
	Port          string
	IP            string
	ServiceName   string
	Host          string
	ServiceWeight uint16
	TTL           uint32
	Priority      uint16
	RelayIP       string
	RelayPort     string
	ListenAddr    string
	LocalPort     string
	LocalRelayIP  string
	UID           string
}

type Option func(drop *Godrop)

const (
	StrategyMDNS = "mdns"
	StrategyHP   = "tcpholepunch"
)

//NewGodrop returns a new godrop server
func NewGodrop(opt ...Option) (*Godrop, error) {
	//default IP.
	myIP, err := getMyIpv4Addr()
	if err != nil {
		return nil, err
	}

	//defafults
	drop := &Godrop{
		Port:          "3000",
		IP:            myIP.String(),
		ServiceName:   "_godrop._tcp.local",
		Host:          "godrop.local",
		ServiceWeight: 0,
		TTL:           0,
		Priority:      0,
		RelayIP:       "127.0.0.1",
		RelayPort:     "7000",
		ListenAddr:    myIP.String(),
		LocalPort:     "4000",
		LocalRelayIP:  myIP.String(),
		UID:           "godrop",
	}

	//override defaults
	for _, option := range opt {
		option(drop)
	}

	// set up tcp server
	server := &server{
		Port: drop.Port,
		IP:   myIP.String(),
	}

	drop.tcpServer = server

	return drop, nil

}

func (drop *Godrop) NewP2PConn(strategy string) ConnectionStrategy {
	s := strings.ToLower(strategy)

	var connStrategy ConnectionStrategy

	switch s {
	case StrategyMDNS:
		connStrategy = Mdns{
			tcpServer:     drop.tcpServer,
			Port:          drop.Port,
			IP:            drop.IP,
			ServiceName:   drop.ServiceName,
			Host:          drop.Host,
			ServiceWeight: drop.ServiceWeight,
			TTL:           drop.TTL,
			Priority:      drop.Priority,
		}
	case StrategyHP:
		connStrategy = TcpHolepunch{
			RelayIP:      drop.RelayIP,
			RelayPort:    drop.RelayPort,
			ListenAddr:   drop.ListenAddr,
			LocalPort:    drop.LocalPort,
			LocalRelayIP: drop.LocalRelayIP,
			UID:          drop.UID,
		}

	}

	return connStrategy
}
