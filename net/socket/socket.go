package socket

import (
	"fmt"
	"net"
	"strconv"

	instanceNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_instance_network"
	ipNameLookup "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_ip_name_lookup"
	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
	wasiTcpCreate "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp_create_socket"
)

func Dial(network string, address string) (net.Conn, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", address, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	n := instanceNetwork.InstanceNetwork()

	remoteAddr, err := resolveAddress(n, host, uint16(port))
	if err != nil {
		return nil, err
	}

	family := addressFamily(remoteAddr)
	socketRes := wasiTcpCreate.CreateTcpSocket(family)
	if socketRes.IsErr() {
		return nil, wasiErrorToGoError(socketRes.Err())
	}
	sock := socketRes.Ok()

	connectRes := sock.StartConnect(n, remoteAddr)
	if connectRes.IsErr() {
		sock.Drop()
		return nil, wasiErrorToGoError(connectRes.Err())
	}

	for {
		pollable := sock.Subscribe()
		pollable.Block()
		pollable.Drop()

		finishRes := sock.FinishConnect()
		if finishRes.IsErr() {
			code := finishRes.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, wasiErrorToGoError(code)
		}
		streams := finishRes.Ok()
		return &wasiConn{
			socket: sock,
			reader: streams.F0,
			writer: streams.F1,
		}, nil
	}
}

func Listen(network string, address string) (net.Listener, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", address, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	if host == "" {
		host = "0.0.0.0"
	}

	n := instanceNetwork.InstanceNetwork()

	localAddr, err := resolveAddress(n, host, uint16(port))
	if err != nil {
		return nil, err
	}

	family := addressFamily(localAddr)
	socketRes := wasiTcpCreate.CreateTcpSocket(family)
	if socketRes.IsErr() {
		return nil, wasiErrorToGoError(socketRes.Err())
	}
	sock := socketRes.Ok()

	bindRes := sock.StartBind(n, localAddr)
	if bindRes.IsErr() {
		sock.Drop()
		return nil, wasiErrorToGoError(bindRes.Err())
	}

	for {
		pollable := sock.Subscribe()
		pollable.Block()
		pollable.Drop()

		finishRes := sock.FinishBind()
		if finishRes.IsErr() {
			code := finishRes.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, wasiErrorToGoError(code)
		}
		break
	}

	listenRes := sock.StartListen()
	if listenRes.IsErr() {
		sock.Drop()
		return nil, wasiErrorToGoError(listenRes.Err())
	}

	for {
		pollable := sock.Subscribe()
		pollable.Block()
		pollable.Drop()

		finishRes := sock.FinishListen()
		if finishRes.IsErr() {
			code := finishRes.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, wasiErrorToGoError(code)
		}
		break
	}

	addrRes := sock.LocalAddress()
	if addrRes.IsErr() {
		sock.Drop()
		return nil, wasiErrorToGoError(addrRes.Err())
	}

	return &wasiListener{
		socket:  sock,
		network: n,
		addr:    ipSocketAddressToNetAddr(addrRes.Ok()),
	}, nil
}

// resolveAddress resolves a host:port into a WASI IpSocketAddress.
// For hostnames, it uses WASI IP name lookup to resolve the first address.
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

func addressFamily(addr wasiNetwork.IpSocketAddress) wasiNetwork.IpAddressFamily {
	if addr.Tag() == wasiNetwork.IpSocketAddressIpv4 {
		return wasiNetwork.IpAddressFamilyIpv4
	}
	return wasiNetwork.IpAddressFamilyIpv6
}

func wasiErrorToGoError(code wasiNetwork.ErrorCode) error {
	switch code {
	case wasiNetwork.ErrorCodeUnknown:
		return fmt.Errorf("unknown error")
	case wasiNetwork.ErrorCodeAccessDenied:
		return fmt.Errorf("access denied")
	case wasiNetwork.ErrorCodeNotSupported:
		return fmt.Errorf("operation not supported")
	case wasiNetwork.ErrorCodeInvalidArgument:
		return fmt.Errorf("invalid argument")
	case wasiNetwork.ErrorCodeOutOfMemory:
		return fmt.Errorf("out of memory")
	case wasiNetwork.ErrorCodeTimeout:
		return fmt.Errorf("operation timed out")
	case wasiNetwork.ErrorCodeConcurrencyConflict:
		return fmt.Errorf("concurrency conflict")
	case wasiNetwork.ErrorCodeNotInProgress:
		return fmt.Errorf("operation not in progress")
	case wasiNetwork.ErrorCodeWouldBlock:
		return fmt.Errorf("operation would block")
	case wasiNetwork.ErrorCodeInvalidState:
		return fmt.Errorf("invalid state")
	case wasiNetwork.ErrorCodeNewSocketLimit:
		return fmt.Errorf("new socket limit reached")
	case wasiNetwork.ErrorCodeAddressNotBindable:
		return fmt.Errorf("address not bindable")
	case wasiNetwork.ErrorCodeAddressInUse:
		return fmt.Errorf("address already in use")
	case wasiNetwork.ErrorCodeRemoteUnreachable:
		return fmt.Errorf("remote unreachable")
	case wasiNetwork.ErrorCodeConnectionRefused:
		return fmt.Errorf("connection refused")
	case wasiNetwork.ErrorCodeConnectionReset:
		return fmt.Errorf("connection reset")
	case wasiNetwork.ErrorCodeConnectionAborted:
		return fmt.Errorf("connection aborted")
	case wasiNetwork.ErrorCodeDatagramTooLarge:
		return fmt.Errorf("datagram too large")
	case wasiNetwork.ErrorCodeNameUnresolvable:
		return fmt.Errorf("name unresolvable")
	case wasiNetwork.ErrorCodeTemporaryResolverFailure:
		return fmt.Errorf("temporary resolver failure")
	case wasiNetwork.ErrorCodePermanentResolverFailure:
		return fmt.Errorf("permanent resolver failure")
	default:
		return fmt.Errorf("wasi socket error: %d", code)
	}
}

func ipSocketAddressToNetAddr(addr wasiNetwork.IpSocketAddress) net.Addr {
	if addr.Tag() == wasiNetwork.IpSocketAddressIpv4 {
		v4 := addr.Ipv4()
		return &net.TCPAddr{
			IP:   net.IPv4(v4.Address.F0, v4.Address.F1, v4.Address.F2, v4.Address.F3),
			Port: int(v4.Port),
		}
	}
	v6 := addr.Ipv6()
	ip := make(net.IP, 16)
	ip[0] = byte(v6.Address.F0 >> 8)
	ip[1] = byte(v6.Address.F0)
	ip[2] = byte(v6.Address.F1 >> 8)
	ip[3] = byte(v6.Address.F1)
	ip[4] = byte(v6.Address.F2 >> 8)
	ip[5] = byte(v6.Address.F2)
	ip[6] = byte(v6.Address.F3 >> 8)
	ip[7] = byte(v6.Address.F3)
	ip[8] = byte(v6.Address.F4 >> 8)
	ip[9] = byte(v6.Address.F4)
	ip[10] = byte(v6.Address.F5 >> 8)
	ip[11] = byte(v6.Address.F5)
	ip[12] = byte(v6.Address.F6 >> 8)
	ip[13] = byte(v6.Address.F6)
	ip[14] = byte(v6.Address.F7 >> 8)
	ip[15] = byte(v6.Address.F7)
	return &net.TCPAddr{
		IP:   ip,
		Port: int(v6.Port),
	}
}
