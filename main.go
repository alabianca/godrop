package main

import (
	"bufio"
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
			result = append(result, buf[:n]...)
		}
	}(quitChan)

	return quitChan
}

func readFromPeer(conn *P2PConn) chan []byte {
	result := make(chan []byte)
	go func(quit chan []byte) {
		buf := make([]byte, 1024)
		data := make([]byte, 0)

		for {
			n, err := conn.Read(buf)

			if err != nil {
				if err == io.EOF {
					data = append(data, buf[:n]...)
					result <- data

				}
				close(result)
				return
			}

			data = append(data, buf[:n]...)
		}

	}(result)

	return result
}

func pipe(stdinChan, fromPeerChan <-chan []byte, conn *P2PConn) {

	for {
		select {
		case data := <-stdinChan:
			//write to the peer
			writeToPeer(data, conn)
			return

		case data := <-fromPeerChan:
			//write to stdout
			writeToStdout(data)
			return
		}
	}
}

func writeToPeer(data []byte, conn *P2PConn) {
	conn.Write(data)
	conn.Close()
}

func writeToStdout(data []byte) {
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(data)
	writer.Flush()
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
	p2pConn := drop.NewP2PConn()

	pipe(readStdin(), readFromPeer(p2pConn), p2pConn)

}
