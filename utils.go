package godrop

import (
	"net"
	"strconv"
	"strings"
)

type NoIPv4FoundError struct{}

func (e NoIPv4FoundError) Error() string {
	return "No IPv4 Interface found"
}

func addressStringToIP(address string) net.IP {
	split := strings.Split(address, "/")
	ipSlice := strings.Split(split[0], ".")

	parts := make([]byte, 4)

	for i := range ipSlice {
		part, _ := strconv.ParseInt(ipSlice[i], 10, 16)
		parts[i] = byte(part)
	}

	ip := net.IPv4(parts[0], parts[1], parts[2], parts[3])

	return ip

}

func getMyIpv4Addr() (net.IP, error) {
	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {

		addr, _ := iface.Addrs()

		for _, a := range addr {
			if strings.Contains(a.String(), ":") { //must be an ipv6
				continue
			}

			ip := addressStringToIP(a.String())

			if ip.IsLoopback() {
				continue
			}

			return ip, nil

		}
	}

	return nil, NoIPv4FoundError{}
}

func fillString(current string, toLength int) string {

	for {
		if len(current) < toLength {
			current = current + "/"
			continue
		}

		break
	}

	return current
}
