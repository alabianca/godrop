package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alabianca/dnsPacket"
)

type Peer struct {
	Port uint16
	IP   string
}

func main() {
	other := Peer{}
	addr, _ := net.ResolveUDPAddr("udp", MulticastAddress)
	//addr, _ := net.ResolveUDPAddr("udp", ":7777")
	pc, err := net.ListenMulticastUDP("udp4", nil, addr)
	//pc, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer pc.Close()

	dnspkt := dnsPacket.DNSPacket{
		Type:    "query",
		ID:      1,
		Opcode:  0,
		Flags:   dnsPacket.FlagsRecurionDesired,
		Qdcount: 1,
	}
	dnspkt.AddQuestion(ServiceName, 1, 33)
	packet := dnsPacket.Encode(&dnspkt)

	srvQueryChan := make(chan dnsPacket.DNSPacket)
	srvResponseChan := make(chan dnsPacket.RecordTypeSRV)
	tcpConnectChan := make(chan dnsPacket.RecordTypeA)
	scanChan := make(chan int)
	buffer := make([]byte, 1024)

	//google, _ := net.ResolveUDPAddr("udp", "8.8.8.8:53")
	conn, _ := net.DialUDP("udp4", nil, addr)

	go func(srvQuery chan dnsPacket.DNSPacket, srvResponse chan dnsPacket.RecordTypeSRV, tcpConnect chan dnsPacket.RecordTypeA, scan chan int) {
		for {
			_, peer, err := pc.ReadFromUDP(buffer)

			if err != nil {
				os.Exit(1)
			}

			decoded := dnsPacket.Decode(buffer)
			//fmt.Println(decoded)
			//fmt.Printf("Got a packet: %s", decoded.Type)
			if decoded.Type == "query" {
				//fmt.Printf("Query From: %s \n", peer)
				responsePacket, ok := hanldeQuery(*peer, *decoded)

				if ok {
					srvQuery <- *responsePacket
				}

			}

			if decoded.Type == "response" {
				packetProcessor, t, ok := getPacketProcessor(*peer, *decoded)

				if !ok {
					scanChan <- 1
				}
				switch t {
				case dnsPacket.DNSRecordTypeSRV:
					srvResponse <- *packetProcessor.(*dnsPacket.RecordTypeSRV)
				case dnsPacket.DNSRecordTypeA:
					tcpConnectChan <- *packetProcessor.(*dnsPacket.RecordTypeA)
				}
			}
			//fmt.Println(decoded)
		}
	}(srvQueryChan, srvResponseChan, tcpConnectChan, scanChan)

	for {

		select {
		case q := <-srvQueryChan:
			conn.Write(dnsPacket.Encode(&q))
			//os.Exit(1)
		case srv := <-srvResponseChan:
			fmt.Println("Got a response: ", srv.Target)
			fmt.Printf("Port: %d\n", srv.Port)
			other.Port = srv.Port
			//now send an 'A' record query to get IP
			query := dnsPacket.DNSPacket{
				Type:    "query",
				ID:      1,
				Opcode:  0,
				Flags:   dnsPacket.FlagsRecurionDesired,
				Qdcount: 1,
			}
			query.AddQuestion(srv.Target, 1, 1)
			conn.Write(dnsPacket.Encode(&query))
		case tcp := <-tcpConnectChan:
			fmt.Printf("this is where I would connect ... %s\n", tcp.IPv4)
			fmt.Println(tcp.IPv4)
			other.IP = tcp.IPv4
		case <-scanChan:
			conn.Write(packet)
		default:
			conn.Write(packet)
			time.Sleep(time.Second * 1)

		}
	}

}

func getPacketProcessor(peer net.UDPAddr, packet dnsPacket.DNSPacket) (dnsPacket.PacketProcessor, int, bool) {
	//1. figure out what type of response it is

	me, err := getMyIpv4Addr()

	if err != nil || peer.IP.Equal(me) {
		return nil, 0, false
	}

	if packet.Ancount <= 0 { //no answers to work with
		return nil, 0, false
	}

	//just looking at the first answer
	answer := packet.Answers[0]

	var t int
	switch answer.Type {
	case dnsPacket.DNSRecordTypeSRV:
		t = dnsPacket.DNSRecordTypeSRV
	case dnsPacket.DNSRecordTypeA:
		t = dnsPacket.DNSRecordTypeA
	}
	packetProcessor := answer.Process()

	return packetProcessor, t, true
}

func hanldeQuery(peer net.UDPAddr, packet dnsPacket.DNSPacket) (*dnsPacket.DNSPacket, bool) {
	//1. check if query is from actual peer and not just from me ...
	me, err := getMyIpv4Addr()
	if err != nil || peer.IP.Equal(me) {
		return nil, false
	}
	//no questions... just drop it
	if packet.Qdcount <= 0 {
		return nil, false
	}
	newPacket := dnsPacket.DNSPacket{
		Type:    "response",
		ID:      packet.ID,
		Opcode:  packet.Opcode,
		Flags:   packet.Flags,
		Qdcount: 1,
	}
	//2. check what type of query - i only look at the first question
	question := packet.Questions[0]

	newPacket.AddQuestion(question.Qname, question.Qclass, question.Qtype)

	switch question.Qtype {
	case dnsPacket.DNSRecordTypeSRV:
		//answer SRV query
		if question.Qname != ServiceName {
			return nil, false
		}

		data := answerSRVQuery()
		newPacket.AddAnswer(ServiceName, 1, 33, 500, len(data), data)
	case dnsPacket.DNSRecordTypeA:
		//answer A query
		if question.Qname != "godrop.local" {
			return nil, false
		}

		data := answerAQuery(me)
		newPacket.AddAnswer("godrop.local", 1, 1, 500, len(data), data)
	}

	newPacket.Type = "response"
	newPacket.Ancount = uint16(len(newPacket.Answers))

	return &newPacket, true
}

func answerDefaultQuery() []byte {
	defaultRercord := dnsPacket.RecordTypeDefault{}

	return defaultRercord.Encode()
}

func answerSRVQuery() []byte {
	srvRecord := dnsPacket.RecordTypeSRV{
		Priority: 0,
		Weight:   0,
		Port:     7777,
		Target:   "godrop.local",
	}

	return srvRecord.Encode()

}

func answerAQuery(me net.IP) []byte {
	aRecord := dnsPacket.RecordTypeA{
		IPv4: me.String(),
	}

	return aRecord.Encode()
}
