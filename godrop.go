package godrop

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

type Godrop struct {
	tcpServer        *Server
	Port             int
	IP               net.IP
	ServiceName      string
	Host             string
	ServiceWeight    uint16
	TTL              uint32
	Priority         uint16
	UID              string
	SharePath        string
	RootCaCert       []byte
	GodropCert       []byte
	GodropPrivateKey []byte
	tlsConfig        *tls.Config
}

type Option func(drop *Godrop)

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
		SharePath:     "",
	}

	//override defaults
	for _, option := range opt {
		option(drop)
	}

	// if tls is desired set up the root certificate as a trusted cert
	if len(drop.RootCaCert) > 0 {
		// assume tls is desired
		drop.tlsConfig = new(tls.Config)
		tlsRoots := x509.NewCertPool()
		if ok := tlsRoots.AppendCertsFromPEM(drop.RootCaCert); !ok {
			return nil, fmt.Errorf("Could Not Parse Provided Root Certificate")
		}

		drop.tlsConfig.RootCAs = tlsRoots
	}

	// set up tcp server
	server := &Server{
		Port:      drop.Port,
		IP:        drop.IP.String(),
		sharePath: drop.SharePath,
		shutdown:  make(chan struct{}),
	}

	if len(drop.GodropCert) > 0 && len(drop.GodropPrivateKey) > 0 {
		// add server (godropCert) certificate and private key to the server
		if cer, err := tls.X509KeyPair(drop.GodropCert, drop.GodropPrivateKey); err != nil {
			return nil, err
		} else {
			server.tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
		}
	}

	drop.tcpServer = server
	drop.Host = drop.UID + "." + drop.Host // root.godrop.local

	return drop, nil

}

// NewMDNSService registers a new godrop service
func (drop *Godrop) NewMDNSService(sharePath string) (*Server, error) {

	meta := []string{
		"version=1.0",
		"name=godrop",
		"uid=" + drop.UID,
		"droplet=" + drop.Host,
	}
	server, err := zeroconf.Register(drop.UID, drop.ServiceName, "local.", drop.Port, meta, nil)

	if err != nil {
		return nil, err
	}

	file, err := os.Open(sharePath)

	defer file.Close()

	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {
		return nil, err
	}

	drop.tcpServer.mdnsService = server
	drop.tcpServer.sharePath = sharePath // todo: some validation it is a valid path
	drop.tcpServer.fInfo = fileInfo

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
func (drop *Godrop) Connect(instance string) (*Session, error) {
	service, err := drop.Lookup(instance)

	if drop.tlsConfig != nil {
		drop.tlsConfig.ServerName = strings.Split(service.Text[2], "=")[1] + "." + strings.Split(service.Text[3], "=")[1] // the drop.Host property of the remote godrop
		fmt.Println("Server Name: ", drop.tlsConfig.ServerName)
	}

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

		conn, err := drop.connect("tcp4", net.JoinHostPort(ip.String(), port))
		fmt.Println(err)
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
			conn, err := drop.connect("tcp6", net.JoinHostPort(ip.String(), port))
			fmt.Println(err)
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

	var encryptionStatus = false
	if drop.tlsConfig != nil {
		encryptionStatus = true
	}

	sesh, err := NewSession(c, true, encryptionStatus)
	sesh.RemoteHost = service.HostName
	sesh.RemoteService = service.ServiceInstanceName()
	sesh.LocalHost = drop.Host
	sesh.RemoteDroplet = strings.Split(service.Text[3], "=")[1]

	if err != nil {
		return nil, err
	}

	return sesh, nil

}

// Establish the underlying tcp connection. If the drop instance's tls config is non nil connect
// attempts to establish a tls connection
func (drop *Godrop) connect(dialType, joinedHostPort string) (net.Conn, error) {
	if drop.tlsConfig == nil {
		return net.Dial(dialType, joinedHostPort)
	}

	//Attempt to establish a TLS connection
	fmt.Println("Attempting TLS Connection")
	return tls.Dial(dialType, joinedHostPort, drop.tlsConfig)
}
