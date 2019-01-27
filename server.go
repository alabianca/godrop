package godrop

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	IP          string
	mdnsService *zeroconf.Server
	shutdown    chan struct{}
}

func (s *Server) Shutdown() {
	fmt.Println("Shutting down...")
	close(s.shutdown)
}

func (s *Server) listen() {
	port := strconv.Itoa(s.Port)
	address := net.JoinHostPort(s.IP, port)

	l, err := net.Listen("tcp4", address)

	if err != nil {
		os.Exit(1)
	}

	connChan := make(chan net.Conn)

	go func(listener net.Listener, c chan net.Conn) {
		conn, err := l.Accept()

		if err != nil {
			return
		}

		c <- conn

	}(l, connChan)

	for {
		select {
		case <-s.shutdown:
			log.Println("shutting down mdns service")
			s.mdnsService.Shutdown()
			return
		case c := <-connChan:
			go s.handleConnection(&c)
		}
	}

}

func (s *Server) handleConnection(conn *net.Conn) {
	fmt.Println("Got a connection ", conn)
}

func mainLoop(s *Server) {
	s.listen()
}
