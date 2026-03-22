package socket

import (
	"fmt"
	"net"

	ipNameLookup "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_ip_name_lookup"
	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
)

// resolveAddress resolves a host:port into a WASI IpSocketAddress.
// For IP literals the address is used directly. For hostnames, WASI IP name
// lookup is used and the first address returned is used — no fallback to
// subsequent addresses is attempted.
func resolveAddress(n *wasiNetwork.Network, host string, port uint16) (wasiNetwork.IpSocketAddress, error) {
	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			return wasiNetwork.MakeIpSocketAddressIpv4(wasiNetwork.Ipv4SocketAddress{
				Port:    port,
				Address: wasiNetwork.Ipv4Address{F0: ip4[0], F1: ip4[1], F2: ip4[2], F3: ip4[3]},
			}), nil
		}
		ip6 := ip.To16()
		return wasiNetwork.MakeIpSocketAddressIpv6(wasiNetwork.Ipv6SocketAddress{
			Port: port,
			Address: wasiNetwork.Ipv6Address{
				F0: uint16(ip6[0])<<8 | uint16(ip6[1]),
				F1: uint16(ip6[2])<<8 | uint16(ip6[3]),
				F2: uint16(ip6[4])<<8 | uint16(ip6[5]),
				F3: uint16(ip6[6])<<8 | uint16(ip6[7]),
				F4: uint16(ip6[8])<<8 | uint16(ip6[9]),
				F5: uint16(ip6[10])<<8 | uint16(ip6[11]),
				F6: uint16(ip6[12])<<8 | uint16(ip6[13]),
				F7: uint16(ip6[14])<<8 | uint16(ip6[15]),
			},
		}), nil
	}

	streamRes := ipNameLookup.ResolveAddresses(n, host)
	if streamRes.IsErr() {
		return wasiNetwork.IpSocketAddress{}, wasiErrorToGoError(streamRes.Err())
	}
	stream := streamRes.Ok()
	defer stream.Drop()

	for {
		pollable := stream.Subscribe()
		pollable.Block()
		pollable.Drop()

		addrRes := stream.ResolveNextAddress()
		if addrRes.IsErr() {
			return wasiNetwork.IpSocketAddress{}, wasiErrorToGoError(addrRes.Err())
		}

		opt := addrRes.Ok()
		if opt.IsNone() {
			break
		}

		addr := opt.Some()
		if addr.Tag() == wasiNetwork.IpAddressIpv4 {
			return wasiNetwork.MakeIpSocketAddressIpv4(wasiNetwork.Ipv4SocketAddress{
				Port:    port,
				Address: addr.Ipv4(),
			}), nil
		}
		if addr.Tag() == wasiNetwork.IpAddressIpv6 {
			return wasiNetwork.MakeIpSocketAddressIpv6(wasiNetwork.Ipv6SocketAddress{
				Port:    port,
				Address: addr.Ipv6(),
			}), nil
		}
	}

	return wasiNetwork.IpSocketAddress{}, fmt.Errorf("could not resolve host %q", host)
}
