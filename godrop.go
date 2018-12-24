package main

import (
	"fmt"
	"time"

	"github.com/alabianca/dnsPacket"
	"github.com/alabianca/mdns"
)

type godrop struct {
	tcpServer     *Server
	peerPort      uint16
	peerIP        string
	stopQueryChan chan int
}

func NewGodrop(conf config) *godrop {

	server := &Server{
		Port: conf.Port,
		IP:   conf.IP,
	}

	drop := godrop{
		tcpServer:     server,
		stopQueryChan: make(chan int),
	}

	mdnsServer, _ := mdns.New()

	mdnsServer.Browse()

	timer := time.NewTicker(time.Duration(3) * time.Second)

	//start out by querying for SRV recorcds
	queryName := conf.ServiceName
	queryType := "SRV"

	go func() {
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
						drop.peerIP = ip
					}
					if port != 0 {
						drop.peerPort = port
					}

					fmt.Println("Ok ", drop)
				}

				if drop.peerIP != "" && drop.peerPort != 0 {
					timer.Stop()
					drop.stopQueryChan <- 1
					return
				}

			case <-timer.C:
				mdnsServer.Query(queryName, "IN", queryType)
			}
		}
	}()

	return &drop

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
