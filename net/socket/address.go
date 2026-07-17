package socket

import (
	"fmt"
	"net"

	instanceNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_instance_network"
	ipNameLookup "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_ip_name_lookup"
	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
)

// resolveHostAddresses resolves host to its WASI IP addresses, invoking visit for
// each one in order. If visit returns true, resolution stops early and any
// remaining addresses are not resolved. IP literals are resolved directly
// without going through the WASI resolver.
func resolveHostAddresses(n *wasiNetwork.Network, host string, visit func(wasiNetwork.IpAddress) bool) error {
	if ip := net.ParseIP(host); ip != nil {
		visit(mapIPToIpAddress(ip))
		return nil
	}

	streamRes := ipNameLookup.ResolveAddresses(n, host)
	if streamRes.IsErr() {
		return mapErrorCode(streamRes.Err())
	}
	stream := streamRes.Ok()
	defer stream.Drop()

	for {
		pollable.AwaitAndDrop(stream.Subscribe())

		addrRes := stream.ResolveNextAddress()
		if addrRes.IsErr() {
			return mapErrorCode(addrRes.Err())
		}

		opt := addrRes.Ok()
		if opt.IsNone() {
			return nil
		}

		if visit(opt.Some()) {
			return nil
		}
	}
}

// resolveAddress resolves a host:port into a WASI IpSocketAddress.
// For IP literals the address is used directly. For hostnames, WASI IP name
// lookup is used and the first address returned is used — no fallback to
// subsequent addresses is attempted.
func resolveAddress(n *wasiNetwork.Network, host string, port uint16) (wasiNetwork.IpSocketAddress, error) {
	var result wasiNetwork.IpSocketAddress
	found := false

	err := resolveHostAddresses(n, host, func(addr wasiNetwork.IpAddress) bool {
		found = true
		if addr.Tag() == wasiNetwork.IpAddressIpv4 {
			result = wasiNetwork.MakeIpSocketAddressIpv4(wasiNetwork.Ipv4SocketAddress{
				Port:    port,
				Address: addr.Ipv4(),
			})
		} else {
			result = wasiNetwork.MakeIpSocketAddressIpv6(wasiNetwork.Ipv6SocketAddress{
				Port:    port,
				Address: addr.Ipv6(),
			})
		}
		return true
	})
	if err != nil {
		return wasiNetwork.IpSocketAddress{}, err
	}
	if !found {
		return wasiNetwork.IpSocketAddress{}, fmt.Errorf("could not resolve host %q", host)
	}
	return result, nil
}

// lookupIPs resolves host to all of its IPv4 and IPv6 addresses using the
// WASI ip-name-lookup interface. IP literals are returned directly without
// going through the resolver.
func lookupIPs(host string) ([]net.IP, error) {
	n := instanceNetwork.InstanceNetwork()

	var ips []net.IP
	err := resolveHostAddresses(n, host, func(addr wasiNetwork.IpAddress) bool {
		ips = append(ips, mapIpAddressToIP(addr))
		return false
	})
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no such host %q", host)
	}
	return ips, nil
}
