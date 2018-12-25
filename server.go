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

func (s *Server) Listen(connectionHandler func(*net.Conn)) {
	address := s.IP + ":" + s.Port

	l, err := net.Listen("tcp4", address)

	if err != nil {
		fmt.Println("Could Not listen")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error Accepting connection")
			os.Exit(1)
			return
		}

		connectionHandler(&conn)
		return
	}

}

func (s *Server) Connect(ip string, port uint16) {
	p := strconv.Itoa(int(port))
	addr := ip + ":" + p

	fmt.Println("Connecting ... ", addr)

	net.Dial("tcp4", addr)
}
