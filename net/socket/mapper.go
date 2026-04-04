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

const byteBits = 8

func mapIPAddress(addr wasiNetwork.IpSocketAddress) net.Addr {
	if addr.Tag() == wasiNetwork.IpSocketAddressIpv4 {
		v4 := addr.Ipv4()
		return &net.TCPAddr{
			IP:   net.IPv4(v4.Address.F0, v4.Address.F1, v4.Address.F2, v4.Address.F3),
			Port: int(v4.Port),
		}
	}
	v6 := addr.Ipv6()
	ip := make(net.IP, net.IPv6len)
	ip[0] = byte(v6.Address.F0 >> byteBits)
	ip[1] = byte(v6.Address.F0) //nolint:gosec // intentional: low byte of uint16
	ip[2] = byte(v6.Address.F1 >> byteBits)
	ip[3] = byte(v6.Address.F1) //nolint:gosec // intentional: low byte of uint16
	ip[4] = byte(v6.Address.F2 >> byteBits)
	ip[5] = byte(v6.Address.F2) //nolint:gosec // intentional: low byte of uint16
	ip[6] = byte(v6.Address.F3 >> byteBits)
	ip[7] = byte(v6.Address.F3) //nolint:gosec // intentional: low byte of uint16
	ip[8] = byte(v6.Address.F4 >> byteBits)
	ip[9] = byte(v6.Address.F4) //nolint:gosec // intentional: low byte of uint16
	ip[10] = byte(v6.Address.F5 >> byteBits)
	ip[11] = byte(v6.Address.F5) //nolint:gosec // intentional: low byte of uint16
	ip[12] = byte(v6.Address.F6 >> byteBits)
	ip[13] = byte(v6.Address.F6) //nolint:gosec // intentional: low byte of uint16
	ip[14] = byte(v6.Address.F7 >> byteBits)
	ip[15] = byte(v6.Address.F7) //nolint:gosec // intentional: low byte of uint16
	return &net.TCPAddr{
		IP:   ip,
		Port: int(v6.Port),
	}
}
