package socket

import (
	"fmt"
	"net"
	"strconv"

	lookup "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_ip_name_lookup"
	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
)

func resolveAddress(address, defaultHost string) (sockets.IpAddress, uint16, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return sockets.IpAddress{}, 0, err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return sockets.IpAddress{}, 0, err
	}
	if host == "" {
		host = defaultHost
	}

	addrs, err := resolveHostAddresses(host)
	if err != nil {
		return sockets.IpAddress{}, 0, err
	}
	if len(addrs) == 0 {
		return sockets.IpAddress{}, 0, fmt.Errorf("resolve %q: no addresses found", host)
	}
	return addrs[0], uint16(port), nil
}

// resolveHostAddresses resolves host to all of its WASI IP addresses. IP
// literals are resolved directly without going through the WASI resolver.
func resolveHostAddresses(host string) ([]sockets.IpAddress, error) {
	if addr, ok := parseIpAddress(host); ok {
		return []sockets.IpAddress{addr}, nil
	}

	resolveRes := lookup.ResolveAddresses(host)
	if resolveRes.IsErr() {
		return nil, mapLookupErrorCode(resolveRes.Err())
	}
	return resolveRes.Ok(), nil
}

// lookupIPs resolves host to all of its IPv4 and IPv6 addresses using the
// WASI ip-name-lookup interface.
func lookupIPs(host string) ([]net.IP, error) {
	addrs, err := resolveHostAddresses(host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no such host %q", host)
	}
	ips := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		ips[i] = mapIP(addr)
	}
	return ips, nil
}
