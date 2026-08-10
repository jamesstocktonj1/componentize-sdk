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

	if addr, ok := parseIpAddress(host); ok {
		return addr, uint16(port), nil
	}

	resolveRes := lookup.ResolveAddresses(host)
	if resolveRes.IsErr() {
		return sockets.IpAddress{}, 0, mapLookupErrorCode(resolveRes.Err())
	}
	addrs := resolveRes.Ok()
	if len(addrs) == 0 {
		return sockets.IpAddress{}, 0, fmt.Errorf("resolve %q: no addresses found", host)
	}
	return addrs[0], uint16(port), nil
}

func mapAddressFamily(addr sockets.IpAddress) sockets.IpAddressFamily {
	if addr.Tag() == sockets.IpAddressIpv4 {
		return sockets.IpAddressFamilyIpv4
	}
	return sockets.IpAddressFamilyIpv6
}

func mapIpSocketAddress(addr sockets.IpAddress, port uint16) sockets.IpSocketAddress {
	switch addr.Tag() {
	case sockets.IpAddressIpv4:
		return sockets.MakeIpSocketAddressIpv4(sockets.Ipv4SocketAddress{
			Port:    port,
			Address: addr.Ipv4(),
		})
	default:
		return sockets.MakeIpSocketAddressIpv6(sockets.Ipv6SocketAddress{
			Port:    port,
			Address: addr.Ipv6(),
		})
	}
}

// parseIpAddress converts a Go IP literal to a WASI IpAddress.
// Returns (addr, true) if host is a valid IP, (zero, false) otherwise.
func parseIpAddress(host string) (sockets.IpAddress, bool) {
	ip := net.ParseIP(host)
	if ip == nil {
		return sockets.IpAddress{}, false
	}
	if ip4 := ip.To4(); ip4 != nil {
		return sockets.MakeIpAddressIpv4(sockets.Ipv4Address{
			F0: ip4[0], F1: ip4[1], F2: ip4[2], F3: ip4[3],
		}), true
	}
	ip6 := ip.To16()
	return sockets.MakeIpAddressIpv6(sockets.Ipv6Address{
		F0: uint16(ip6[0])<<8 | uint16(ip6[1]),
		F1: uint16(ip6[2])<<8 | uint16(ip6[3]),
		F2: uint16(ip6[4])<<8 | uint16(ip6[5]),
		F3: uint16(ip6[6])<<8 | uint16(ip6[7]),
		F4: uint16(ip6[8])<<8 | uint16(ip6[9]),
		F5: uint16(ip6[10])<<8 | uint16(ip6[11]),
		F6: uint16(ip6[12])<<8 | uint16(ip6[13]),
		F7: uint16(ip6[14])<<8 | uint16(ip6[15]),
	}), true
}

func mapNetAddr(addr sockets.IpSocketAddress) net.Addr {
	switch addr.Tag() {
	case sockets.IpSocketAddressIpv4:
		v4 := addr.Ipv4()
		ip := net.IPv4(v4.Address.F0, v4.Address.F1, v4.Address.F2, v4.Address.F3)
		return &net.TCPAddr{IP: ip, Port: int(v4.Port)}
	case sockets.IpSocketAddressIpv6:
		v6 := addr.Ipv6()
		ip := make(net.IP, 16)
		for i, seg := range [8]uint16{
			v6.Address.F0, v6.Address.F1, v6.Address.F2, v6.Address.F3,
			v6.Address.F4, v6.Address.F5, v6.Address.F6, v6.Address.F7,
		} {
			ip[i*2] = byte(seg >> 8)
			ip[i*2+1] = byte(seg)
		}
		return &net.TCPAddr{IP: ip, Port: int(v6.Port)}
	}
	return nil
}
