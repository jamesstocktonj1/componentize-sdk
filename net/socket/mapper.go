package socket

import (
	"fmt"
	"net"

	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
)

func mapAddressFamily(addr wasiNetwork.IpSocketAddress) wasiNetwork.IpAddressFamily {
	if addr.Tag() == wasiNetwork.IpSocketAddressIpv4 {
		return wasiNetwork.IpAddressFamilyIpv4
	}
	return wasiNetwork.IpAddressFamilyIpv6
}

func mapErrorCode(code wasiNetwork.ErrorCode) error {
	switch code {
	case wasiNetwork.ErrorCodeUnknown:
		return ErrUnknown
	case wasiNetwork.ErrorCodeAccessDenied:
		return ErrAccessDenied
	case wasiNetwork.ErrorCodeNotSupported:
		return ErrNotSupported
	case wasiNetwork.ErrorCodeInvalidArgument:
		return ErrInvalidArgument
	case wasiNetwork.ErrorCodeOutOfMemory:
		return ErrOutOfMemory
	case wasiNetwork.ErrorCodeTimeout:
		return ErrTimeout
	case wasiNetwork.ErrorCodeConcurrencyConflict:
		return ErrConcurrencyConflict
	case wasiNetwork.ErrorCodeNotInProgress:
		return ErrNotInProgress
	case wasiNetwork.ErrorCodeWouldBlock:
		return ErrWouldBlock
	case wasiNetwork.ErrorCodeInvalidState:
		return ErrInvalidState
	case wasiNetwork.ErrorCodeNewSocketLimit:
		return ErrNewSocketLimit
	case wasiNetwork.ErrorCodeAddressNotBindable:
		return ErrAddressNotBindable
	case wasiNetwork.ErrorCodeAddressInUse:
		return ErrAddressInUse
	case wasiNetwork.ErrorCodeRemoteUnreachable:
		return ErrRemoteUnreachable
	case wasiNetwork.ErrorCodeConnectionRefused:
		return ErrConnectionRefused
	case wasiNetwork.ErrorCodeConnectionReset:
		return ErrConnectionReset
	case wasiNetwork.ErrorCodeConnectionAborted:
		return ErrConnectionAborted
	case wasiNetwork.ErrorCodeDatagramTooLarge:
		return ErrDatagramTooLarge
	case wasiNetwork.ErrorCodeNameUnresolvable:
		return ErrNameUnresolvable
	case wasiNetwork.ErrorCodeTemporaryResolverFailure:
		return ErrTemporaryResolverFailure
	case wasiNetwork.ErrorCodePermanentResolverFailure:
		return ErrPermanentResolverFailure
	default:
		return fmt.Errorf("wasi socket error: %d", code)
	}
}

func mapIpAddressToIP(addr wasiNetwork.IpAddress) net.IP {
	if addr.Tag() == wasiNetwork.IpAddressIpv4 {
		v4 := addr.Ipv4()
		return net.IPv4(v4.F0, v4.F1, v4.F2, v4.F3)
	}
	v6 := addr.Ipv6()
	ip := make(net.IP, 16)
	ip[0] = byte(v6.F0 >> 8)
	ip[1] = byte(v6.F0)
	ip[2] = byte(v6.F1 >> 8)
	ip[3] = byte(v6.F1)
	ip[4] = byte(v6.F2 >> 8)
	ip[5] = byte(v6.F2)
	ip[6] = byte(v6.F3 >> 8)
	ip[7] = byte(v6.F3)
	ip[8] = byte(v6.F4 >> 8)
	ip[9] = byte(v6.F4)
	ip[10] = byte(v6.F5 >> 8)
	ip[11] = byte(v6.F5)
	ip[12] = byte(v6.F6 >> 8)
	ip[13] = byte(v6.F6)
	ip[14] = byte(v6.F7 >> 8)
	ip[15] = byte(v6.F7)
	return ip
}

func mapIPToIpAddress(ip net.IP) wasiNetwork.IpAddress {
	if ip4 := ip.To4(); ip4 != nil {
		return wasiNetwork.MakeIpAddressIpv4(wasiNetwork.Ipv4Address{F0: ip4[0], F1: ip4[1], F2: ip4[2], F3: ip4[3]})
	}
	ip6 := ip.To16()
	return wasiNetwork.MakeIpAddressIpv6(wasiNetwork.Ipv6Address{
		F0: uint16(ip6[0])<<8 | uint16(ip6[1]),
		F1: uint16(ip6[2])<<8 | uint16(ip6[3]),
		F2: uint16(ip6[4])<<8 | uint16(ip6[5]),
		F3: uint16(ip6[6])<<8 | uint16(ip6[7]),
		F4: uint16(ip6[8])<<8 | uint16(ip6[9]),
		F5: uint16(ip6[10])<<8 | uint16(ip6[11]),
		F6: uint16(ip6[12])<<8 | uint16(ip6[13]),
		F7: uint16(ip6[14])<<8 | uint16(ip6[15]),
	})
}

func mapIpAddress(addr wasiNetwork.IpSocketAddress) net.Addr {
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
