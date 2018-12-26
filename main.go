package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func readStdin() chan []byte {

	quitChan := make(chan []byte)

	go func(quit chan []byte) {
		buf := make([]byte, 1024)
		reader := bufio.NewReader(os.Stdin)
		result := make([]byte, 0)

		for {
			n, err := reader.Read(buf)

			if err != nil {
				if err == io.EOF {
					result = append(result, buf[:n]...)
					quit <- result
					close(quit)
					return
				}
			}
			fmt.Println("Appending to result")
			result = append(result, buf[:n]...)
		}
	}(quitChan)

	return quitChan
}

func writeToPeer(start <-chan []byte, conn *P2PConn) {

	go func() {
		fmt.Println("In write to peer")
		data := <-start
		fmt.Println("In write to peer after unlocking channel")
		conn.Write(data)
	}()
}

func readFromPeer(conn *P2PConn) chan []byte {
	result := make(chan []byte)
	fmt.Println("In read from peer")
	go func(quit chan []byte) {
		buf := make([]byte, 1024)
		data := make([]byte, 0)

		for {
			n, err := conn.Read(buf)

			if err != nil {
				if err == io.EOF {
					data = append(data, buf[:n]...)
					result <- data
					close(result)
					return
				}
			}
			fmt.Println("reading from peer")
			data = append(data, buf[:n]...)
		}

	}(result)

	return result
}

func writeToStdout(start <-chan []byte) {
	data := <-start

	writer := bufio.NewWriter(os.Stdout)

	writer.Write(data)
	writer.Flush()

	return
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
	p2pConn := drop.NewP2PConn(conf)

	writeToPeer(readStdin(), p2pConn)
	writeToStdout(readFromPeer(p2pConn))

	// peerChannel := ScanForPeers(conf)
	// drop := NewGodrop(conf)

	// drop.ReadAll()

	// drop.Listen(func(conn *net.TCPConn) {
	// 	drop.handleConnection(conn)
	// })

	// for {
	// 	select {
	// 	case peer := <-peerChannel:
	// 		drop.peer = &peer

	// 		if conf.IP < drop.peer.IP {
	// 			conn, _ := drop.Connect(peer.IP, peer.Port)
	// 			drop.handleConnection(conn)
	// 		}
	// 	}
	// }
}
