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

	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp4", address)

	if tcpErr != nil {
		fmt.Println("Error")
		os.Exit(1)
	}

	l, err := net.ListenTCP("tcp4", tcpAddr)

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
		tcpConn, _ := conn.(*net.TCPConn)
		connectionHandler(tcpConn)
		return
	}

}

func (s *Server) Connect(ip string, port uint16) (*net.TCPConn, error) {
	p := strconv.Itoa(int(port))
	addr := ip + ":" + p

	fmt.Println("Connecting ... ", addr)
	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp4", addr)

	if tcpAddr != nil {
		return nil, tcpErr
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
