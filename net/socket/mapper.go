package socket

import (
	"fmt"
	"net"

	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
)

func addressFamily(addr wasiNetwork.IpSocketAddress) wasiNetwork.IpAddressFamily {
	if addr.Tag() == wasiNetwork.IpSocketAddressIpv4 {
		return wasiNetwork.IpAddressFamilyIpv4
	}
	return wasiNetwork.IpAddressFamilyIpv6
}

func wasiErrorToGoError(code wasiNetwork.ErrorCode) error {
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
