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

func writeToPeer(start <-chan []byte, conn *P2PConn) {

	go func() {
		data := <-start
		conn.Write(data)
		conn.Close()
	}()
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
					close(result)
					return
				} else {
					close(result)
					return
				}
			}

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

}
