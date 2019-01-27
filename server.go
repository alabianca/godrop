package godrop

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	IP          string
	mdnsService *zeroconf.Server
}

func (s *Server) listen() {
	port := strconv.Itoa(s.Port)
	address := net.JoinHostPort(s.IP, port)

	l, err := net.Listen("tcp4", address)

	if err != nil {
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			os.Exit(1)
			return
		}

		fmt.Println("Got a connection: ", conn.RemoteAddr().String())
		// tcpConn, _ := conn.(*net.TCPConn)
		//connectionHandler(tcpConn)
		return
	}

}

func (s *Server) Connect(ip string, port uint16) (*net.TCPConn, error) {
	p := strconv.Itoa(int(port))
	addr := ip + ":" + p

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	tcpConn, _ := conn.(*net.TCPConn)
	return tcpConn, nil
}

func mainLoop(s *Server) {
	s.listen()
}
