package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/alabianca/dnsPacket"
	"github.com/alabianca/mdns"
)

type Peer struct {
	Port uint16
	IP   string
}

type godrop struct {
	tcpServer *Server
	peer      *Peer
	reader    io.Reader
	writer    io.Writer
	conn      *net.TCPConn
	buf       []byte
}

func (drop *godrop) Listen(connectionHandler func(*net.TCPConn)) {
	go drop.tcpServer.Listen(connectionHandler)
}

func (drop *godrop) Connect(ip string, port uint16) (*net.TCPConn, error) {
	conn, err := drop.tcpServer.Connect(ip, port)

	if err != nil {
		return nil, err
	}

	drop.conn = conn

	return conn, nil
}

func (drop *godrop) Write(data []byte) {
	fmt.Println("Going to write ...")
	drop.writer.Write(data)
}

func (drop *godrop) handleConnection(conn *net.TCPConn) {
	if len(drop.buf) > 0 {
		fmt.Println("Writing into connection")
		conn.Write(drop.buf)
		conn.Close()
		return
	}

	buf := make([]byte, 1024)

	for {
		fmt.Println("Reading from connection")
		n, err := conn.Read(buf)
		fmt.Println("Read from connection")
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF")
				buf = append(buf, buf[:n]...)
				break
			}
		}

		buf = append(buf, buf[:n]...)
	}

	drop.Write(buf)

}

func (drop *godrop) ReadAll() error {
	buf := make([]byte, 1024)

	info, _ := os.Stdin.Stat()

	//no stdin
	if info.Size() <= 0 {
		return nil
	}

	for {
		n, err := drop.reader.Read(buf)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		drop.buf = append(drop.buf, buf[:n]...)

	}
	fmt.Println(drop.buf)
	return nil
}

func NewGodrop(conf config) *godrop {
	server := &Server{
		Port: conf.Port,
		IP:   conf.IP,
	}

	drop := godrop{
		tcpServer: server,
		reader:    bufio.NewReader(os.Stdin),
		writer:    bufio.NewWriter(os.Stdout),
	}

	return &drop

}

func ScanForPeers(conf config) chan Peer {

	peer := Peer{}
	quitChan := make(chan Peer)

	mdnsServer, _ := mdns.New()

	mdnsServer.Browse()

	timer := time.NewTicker(time.Duration(3) * time.Second)

	//start out by querying for SRV recorcds
	queryName := conf.ServiceName
	queryType := "SRV"

	go func() {
		//send a query imediately
		mdnsServer.Query(queryName, "IN", queryType)

		for {
			select {
			case packet := <-mdnsServer.QueryChan:
				responseData, ok := handleQuery(packet, conf.ServiceName, conf.Host)

				if ok {
					name := packet.Questions[0].Qname
					anType := packet.Questions[0].Qtype
					mdnsServer.Respond(name, anType, packet, responseData)
				}
			case packet := <-mdnsServer.ResponseChan:
				//is the response one we care about?
				ok := handleResponse(packet, conf.ServiceName, conf.Host)

				if ok {
					//at this point we got a successfull srv record from a peer. Switch the query mode to 'A' records
					queryName = conf.Host
					queryType = "A"
					ip, port := getPeerData(packet.Answers[0])

					if ip != "" {
						peer.IP = ip
					}
					if port != 0 {
						peer.Port = port
					}
				}

				if peer.IP != "" && peer.Port != 0 {
					timer.Stop()
					quitChan <- peer
				}

			case <-timer.C:
				mdnsServer.Query(queryName, "IN", queryType)
			}
		}
	}()

	return quitChan

}

func getPeerData(dnsResponse dnsPacket.Answer) (string, uint16) {

	var ip string
	var port uint16

	switch dnsResponse.Type {
	//The dns response was an A record
	case 1:
		record := dnsResponse.Process()
		aRecord, _ := record.(*dnsPacket.RecordTypeA)
		ip = aRecord.IPv4

	//The dns response was an SRV record
	case 33:
		record := dnsResponse.Process()
		srvRecord, _ := record.(*dnsPacket.RecordTypeSRV)
		port = srvRecord.Port
	}

	return ip, port
}

func handleResponse(response dnsPacket.DNSPacket, serviceName string, host string) bool {
	//check if the response is as a result from our query
	if response.Ancount <= 0 {
		return false
	}
	answer := response.Answers[0]

	var name string
	switch answer.Type {
	case 1:
		name = host

	case 33:
		name = serviceName
	default:
		name = host
	}

	//if we don't know the name jsut return
	if name != answer.Name {
		return false
	}

	return true

}

func handleQuery(query dnsPacket.DNSPacket, serviceName string, host string) ([]byte, bool) {
	if query.Qdcount <= 0 {
		return nil, false
	}

	myIp, err := getMyIpv4Addr()

	if err != nil {
		return nil, false
	}

	question := query.Questions[0]

	var name string
	switch question.Qtype {
	case 1:
		name = host

	case 33:
		name = serviceName
	default:
		name = host
	}

	if name != question.Qname {
		return nil, false
	}

	var data []byte
	switch question.Qtype {
	case 1:
		responseData := dnsPacket.RecordTypeA{
			IPv4: myIp.String(),
		}

		data = responseData.Encode()
	case 33:
		responseData := dnsPacket.RecordTypeSRV{
			Target:   host,
			Port:     7777,
			Weight:   0,
			Priority: 0,
		}

		data = responseData.Encode()
	}

	return data, true
}
