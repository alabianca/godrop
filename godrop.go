package godrop

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/grandcat/zeroconf"
)

type Godrop struct {
	tcpServer *Server
	//peer          *Peer
	Port          int
	IP            net.IP
	ServiceName   string
	Host          string
	ServiceWeight uint16
	TTL           uint32
	Priority      uint16
	UID           string
}

type Option func(drop *Godrop)

const (
	StrategyMDNS = "mdns"
	StrategyHP   = "tcpholepunch"
)

// NewGodrop returns a new godrop server
func NewGodrop(opt ...Option) (*Godrop, error) {
	//default IP.
	myIP, err := getMyIpv4Addr()
	if err != nil {
		return nil, err
	}

	//defafults
	drop := &Godrop{
		Port:          3000,
		IP:            myIP,
		ServiceName:   "_godrop._tcp",
		Host:          "godrop.local",
		ServiceWeight: 0,
		TTL:           0,
		Priority:      0,
		UID:           "root",
	}

	//override defaults
	for _, option := range opt {
		option(drop)
	}

	// set up tcp server
	server := &Server{
		Port:     drop.Port,
		IP:       drop.IP.String(),
		shutdown: make(chan struct{}),
	}

	drop.tcpServer = server
	drop.Host = drop.UID + "." + drop.Host // root.godrop.local

	return drop, nil

}

// NewMDNSService registers a new godrop service
func (drop *Godrop) NewMDNSService() (*Server, error) {

	meta := []string{
		"version=1.0",
		"name=godrop",
		"uid=" + drop.UID,
	}
	server, err := zeroconf.Register(drop.UID, drop.ServiceName, "local.", drop.Port, meta, nil)

	if err != nil {
		return nil, err
	}

	drop.tcpServer.mdnsService = server

	go mainLoop(drop.tcpServer)

	return drop.tcpServer, nil
}

// Discover browses the local network for _godrop_.tcp instances
// it will browse for the given timeout
func (drop *Godrop) Discover(timeout time.Duration) ([]*zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)

	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	res := make([]*zeroconf.ServiceEntry, 0)
	go func(results <-chan *zeroconf.ServiceEntry) {

		for entry := range results {
			res = append(res, entry)
		}

	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*timeout)

	defer cancel()

	errb := resolver.Browse(ctx, drop.ServiceName, "local.", entries)

	if errb != nil {
		return nil, fmt.Errorf("Error Browsing")
	}

	<-ctx.Done()

	return res, nil
}

// Lookup queries the local network for the given _godrop._tcp instance
func (drop *Godrop) Lookup(instance string) (*zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)

	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	result := make(chan *zeroconf.ServiceEntry)
	go func(results chan *zeroconf.ServiceEntry) {
		entry := <-results

		result <- entry

	}(entries)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	errl := resolver.Lookup(ctx, instance, drop.ServiceName, "local.", entries)

	if errl != nil {
		return nil, errl
	}

	res := <-result

	return res, nil

}

// Connect utilizes godrop.Lookup(instance) to connect to the given instance if found
// godrop will attempt to connect to all ip addresses that are advertised and will use the first
// successfull connection
func (drop *Godrop) Connect(instance string) (net.Conn, error) {
	service, err := drop.Lookup(instance)

	if err != nil {
		return nil, err
	}
	port := strconv.Itoa(service.Port)
	var found bool
	var c net.Conn
	// try all ip addresses to connect to the service
	// start with ipv4 addresses
	for i := 0; i < len(service.AddrIPv4); i++ {
		ip := service.AddrIPv4[i]

		conn, err := net.Dial("tcp4", net.JoinHostPort(ip.String(), port))

		if err == nil {
			c = conn
			found = true
			break
		}
	}

	// if we still have not connected. try all ipv6 addresses
	if !found {
		for i := 0; i < len(service.AddrIPv6); i++ {
			ip := service.AddrIPv6[i]
			conn, err := net.Dial("tcp6", net.JoinHostPort(ip.String(), port))

			if err == nil {
				c = conn
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("Could Not Connect to the service")
	}

	return c, nil

}
