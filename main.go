package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	myIp, err := getMyIpv4Addr()

	if err != nil {
		os.Exit(1)
	}

	conf := config{
		Port:          "7777",
		IP:            myIp.String(),
		ServiceName:   "_godrop._tcp.local",
		Host:          "godrop.local",
		Priority:      0,
		ServiceWeight: 0,
		TTL:           500,
	}

	peerChannel := ScanForPeers(conf)
	drop := NewGodrop(conf)

	drop.ReadAll()

	drop.Listen(func(conn *net.TCPConn) {

	})

	for {
		select {
		case peer := <-peerChannel:
			drop.peer = &peer

			if conf.IP < drop.peer.IP {
				fmt.Println("connect")
				drop.Connect(peer.IP, peer.Port)
			}
		}
	}
}
