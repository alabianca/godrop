package godrop

import (
	"net"
	"strconv"
	"time"

	"github.com/alabianca/dnsPacket"
	"github.com/alabianca/mdns"
)

type Peer struct {
	Port uint16
	IP   string
	Host string
}

type Mdns struct {
	tcpServer     *server
	mdnsServer    *mdns.Server
	peer          *Peer
	Port          string
	IP            string
	ServiceName   string
	Host          string
	ServiceWeight uint16
	TTL           uint32
	Priority      uint16
}

func (m *Mdns) listen(connectionHandler func(*net.TCPConn)) {
	go m.tcpServer.Listen(connectionHandler)
}

func (m *Mdns) connect(ip string, port uint16) (*net.TCPConn, error) {
	conn, err := m.tcpServer.Connect(ip, port)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Advertise the godrop service by listening to queries and responding to them
func (m *Mdns) Advertise() *P2PConn {

	quitChan := make(chan P2PConn)

	m.mdnsServer.Browse()

	m.listen(func(conn *net.TCPConn) {
		c := P2PConn{
			Conn: conn,
		}
		quitChan <- c
	})

	go func() {

		for {
			select {
			case packet := <-m.mdnsServer.QueryChan:
				responseData, ok := handleQuery(packet, m.ServiceName, m.Host, m.Port)
				if ok {
					name := packet.Questions[0].Qname
					anType := packet.Questions[0].Qtype
					m.mdnsServer.Respond(name, anType, packet, responseData)
				}
			case <-m.mdnsServer.ResponseChan:
				//nothing to do
			}
		}
	}()

	p2pConn := <-quitChan
	return &p2pConn

}

// Browse for godrop services in the local network by sending a SRV query every 3 seconds
func (m *Mdns) Browse(onPeerDiscovered func(peer Peer)) {

	m.mdnsServer.Browse()

	timer := time.NewTicker(time.Duration(3) * time.Second)

	//start out by querying for SRV recorcds
	queryName := m.ServiceName
	queryType := "SRV"

	go func() {
		//send a query imediately
		m.mdnsServer.Query(queryName, "IN", queryType)

		for {
			select {
			case <-m.mdnsServer.QueryChan:

			case packet := <-m.mdnsServer.ResponseChan:
				//is the response one we care about?
				t, ok := handleResponse(packet, m.ServiceName, m.Host)

				if ok && t == 33 { //only if SRV record
					//at this point we got a successfull srv record from a peer. Switch the query mode to 'A' records
					answer := packet.Answers[0].Process()
					srvRecord := answer.(*dnsPacket.RecordTypeSRV)

					peer := Peer{Port: srvRecord.Port, Host: srvRecord.Target}

					onPeerDiscovered(peer)
				}

			case <-timer.C:
				m.mdnsServer.Query(queryName, "IN", queryType)
			}
		}
	}()
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

func handleResponse(response dnsPacket.DNSPacket, serviceName string, host string) (int, bool) {
	//check if the response is as a result from our query
	if response.Ancount <= 0 {
		return 0, false
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
		return -1, false
	}

	return answer.Type, true

}

func handleQuery(query dnsPacket.DNSPacket, serviceName string, host string, port string) ([]byte, bool) {
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

	p64, _ := strconv.ParseInt(port, 10, 16)
	p16 := uint16(p64)

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
			Port:     p16,
			Weight:   0,
			Priority: 0,
		}

		data = responseData.Encode()
	}

	return data, true
}
