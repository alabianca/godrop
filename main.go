package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alabianca/dnsPacket"
)

func main() {

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
	//test answer
	// d := []byte{0, 0, 0, 0, 219, 84, 3, 49, 50, 55, 1, 48, 1, 48, 1, 49, 0}
	// dnspkt.AddAnswer("airpaste-global", 1, 33, 5, 17, d)
	//dnspkt.AddQuestion("_imaps._tcp.gmail.com", 1, 33)
	packet := dnsPacket.Encode(&dnspkt)

	quitChan := make(chan int)
	buffer := make([]byte, 1024)

	//google, _ := net.ResolveUDPAddr("udp", "8.8.8.8:53")
	conn, _ := net.DialUDP("udp4", nil, addr)

	go func(quit chan int) {
		for {
			_, peer, err := pc.ReadFromUDP(buffer)

			if err != nil {
				os.Exit(1)
			}

			decoded := dnsPacket.Decode(buffer)

			if decoded.Type == "query" {
				fmt.Printf("Query From: %s \n %s", peer, decoded)
				hanldeQuery(*peer, *decoded)
			}

			//fmt.Println(decoded)
			if len(decoded.Answers) > 0 {
				record := decoded.Answers[0].Process()
				t, ok := record.(*dnsPacket.RecordTypeSRV)

				if ok {
					fmt.Println(t.Port)
					fmt.Println(t.Priority)
					fmt.Println(t.Target)
					fmt.Println(t.Weight)
				}

			}
			//fmt.Println(decoded)

			quit <- 1
		}
	}(quitChan)

	for {
		select {
		case q := <-quitChan:
			fmt.Println(q)
			//os.Exit(1)
		default:
			conn.Write(packet)
			time.Sleep(time.Second * 1)

		}
	}

}

func hanldeQuery(peer net.UDPAddr, packet dnsPacket.DNSPacket) {
	//1. check if query is from actual peer and not just from me ...
	me, err := getMyIpv4Addr()
	if err != nil || peer.IP.Equal(me) {
		return
	}
	//no questions... just drop it
	if packet.Qdcount <= 0 {
		return
	}

	//2. check what type of query - i only look at the first question
	question := packet.Questions[0]

	switch question.Qtype {
	case dnsPacket.DNSRecordTypeSRV:
		//answer SRV query
		fmt.Println("Handling an SRV Query")
		data := answerSRVQuery()
		packet.AddAnswer(ServiceName, 1, 33, 500, len(data), data)
	case dnsPacket.DNSRecordTypeA:
		//answer A query
		data := answerAQuery(me)
		packet.AddAnswer("godrop.local", 1, 1, 500, len(data), data)

	}
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
