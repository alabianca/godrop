package main

import (
	"fmt"
	"os"
)

type Peer struct {
	port uint16
	ip   string
}

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

	drop := NewGodrop(conf)

	for {
		select {
		case <-drop.stopQueryChan:
			fmt.Println("Stop Query")
			fmt.Println(drop)
		}
	}
}
