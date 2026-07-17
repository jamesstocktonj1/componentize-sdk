package socket

import (
	"context"
	"fmt"
	"net"
)

// LookupIP looks up host using the local resolver.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	ips, err := lookupIPs(host)
	if err != nil {
		return nil, err
	}

	var filtered []net.IP
	switch network {
	case "ip", "":
		filtered = ips
	case "ip4":
		for _, ip := range ips {
			if ip4 := ip.To4(); ip4 != nil {
				filtered = append(filtered, ip4)
			}
		}
	case "ip6":
		for _, ip := range ips {
			if ip.To4() == nil {
				filtered = append(filtered, ip)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported network %q", network)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no addresses found for host %q on network %q", host, network)
	}
	return filtered, nil
}

// LookupIPAddr looks up host using the local resolver.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	ips, err := lookupIPs(host)
	if err != nil {
		return nil, err
	}

	addrs := make([]net.IPAddr, len(ips))
	for i, ip := range ips {
		addrs[i] = net.IPAddr{IP: ip}
	}
	return addrs, nil
}
