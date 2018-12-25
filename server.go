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

func (s *Server) Listen(connectionHandler func(*net.TCPConn)) {
	address := s.IP + ":" + s.Port

	l, err := net.Listen("tcp4", address)

	if err != nil {
		fmt.Println("Could Not listen")
		os.Exit(1)
	}

	for {
		fmt.Println("Listening for connection on at ", s)
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error Accepting connection")
			os.Exit(1)
			return
		}

		tcpConn, _ := conn.(*net.TCPConn)
		connectionHandler(tcpConn)
		return
	}

}

func (s *Server) Connect(ip string, port uint16) (*net.TCPConn, error) {
	p := strconv.Itoa(int(port))
	addr := ip + ":" + p

	fmt.Println("Connecting ... ", addr)

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	tcpConn, _ := conn.(*net.TCPConn)
	return tcpConn, nil
}
