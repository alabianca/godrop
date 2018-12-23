package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type Server struct {
	Port string
	IP   string
}

func (s *Server) Listen() {
	address := s.IP + ":" + s.Port

	l, err := net.Listen("tcp4", address)

	if err != nil {
		fmt.Println("Could Not listen")
		os.Exit(1)
	}

	for {
		_, err := l.Accept()

		if err != nil {
			fmt.Println("Error Accepting connection")
			os.Exit(1)
			return
		}

		fmt.Println("Got a connection")
	}

}

func (s *Server) Connect(ip string, port uint16) {
	p := strconv.Itoa(int(port))
	addr := ip + ":" + p

	fmt.Println("Connecting ...")

	net.Dial("tcp4", ip+addr)
}
