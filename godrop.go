package main

import (
	"fmt"
	"net"
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
}

func (drop *godrop) Listen(connectionHandler func(*net.Conn)) {
	go drop.tcpServer.Listen(connectionHandler)
}

func (drop *godrop) Connect(ip string, port uint16) {
	drop.tcpServer.Connect(ip, port)
}

func NewGodrop(conf config) *godrop {
	server := &Server{
		Port: conf.Port,
		IP:   conf.IP,
	}

	drop := godrop{
		tcpServer: server,
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
	fmt.Println("Handle Response")
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
