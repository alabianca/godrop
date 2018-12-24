package main

import (
	"fmt"
	"time"

	"github.com/alabianca/dnsPacket"

	"github.com/alabianca/mdns"
)

type Peer struct {
	port uint16
	ip   string
}

const (
	serviceName = "_godrop._tcp.local"
	host        = "godrop.local"
)

func main() {

	mdnsServer, _ := mdns.New()

	mdnsServer.Browse()

	queryName := serviceName
	queryType := "SRV"
	timer := time.NewTicker(time.Duration(5) * time.Second)

	for {
		select {
		case packet := <-mdnsServer.QueryChan:
			fmt.Println("Got a query")
			responseData, ok := handleQuery(packet)

			if ok {
				name := packet.Questions[0].Qname
				anType := packet.Questions[0].Qtype
				mdnsServer.Respond(name, anType, packet, responseData)
			}

		case packet := <-mdnsServer.ResponseChan:
			fmt.Println("got a response")
			ok := handleResponse(packet)

			if ok {
				queryName = host
				queryType = "A"
			}

		case <-timer.C:
			mdnsServer.Query(queryName, "IN", queryType)
		}
	}
}

func handleResponse(response dnsPacket.DNSPacket) bool {
	//check if the response is as a result from our query
	fmt.Println("Handle Response")
	fmt.Println(response)
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

func handleQuery(query dnsPacket.DNSPacket) ([]byte, bool) {
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
