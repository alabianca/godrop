package main

import (
	"fmt"
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

	for {
		select {
		case peer := <-peerChannel:
			fmt.Println("Stop Query")
			fmt.Println(peer)
		}
	}
}
