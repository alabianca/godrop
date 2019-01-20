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
}

type Mdns struct {
	tcpServer     *server
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

func (m Mdns) Connect(peer string) (*P2PConn, error) {
	p := Peer{}

	quitChan := make(chan P2PConn)

	mdnsServer, _ := mdns.New()

	mdnsServer.Browse()

	m.listen(func(conn *net.TCPConn) {
		c := P2PConn{
			Conn: conn,
		}
		quitChan <- c
	})

	timer := time.NewTicker(time.Duration(3) * time.Second)

	//start out by querying for SRV recorcds
	queryName := m.ServiceName
	queryType := "SRV"

	go func() {
		//send a query imediately
		mdnsServer.Query(queryName, "IN", queryType)

		for {
			select {
			case packet := <-mdnsServer.QueryChan:
				responseData, ok := handleQuery(packet, m.ServiceName, m.Host, m.Port)

				if ok {
					name := packet.Questions[0].Qname
					anType := packet.Questions[0].Qtype
					mdnsServer.Respond(name, anType, packet, responseData)
				}
			case packet := <-mdnsServer.ResponseChan:
				//is the response one we care about?
				ok := handleResponse(packet, m.ServiceName, m.Host)

				if ok {
					//at this point we got a successfull srv record from a peer. Switch the query mode to 'A' records
					queryName = m.Host
					queryType = "A"
					ip, port := getPeerData(packet.Answers[0])

					if ip != "" {
						p.IP = ip
					}
					if port != 0 {
						p.Port = port
					}
				}

				if p.IP != "" && p.Port != 0 {
					timer.Stop()

					conn, _ := m.connect(p.IP, p.Port)

					c := P2PConn{
						Conn: conn,
					}

					quitChan <- c
				}

			case <-timer.C:
				mdnsServer.Query(queryName, "IN", queryType)
			}
		}
	}()

	p2pConn := <-quitChan
	return &p2pConn, nil
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
