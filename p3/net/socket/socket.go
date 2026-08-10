package socket

import (
	"fmt"
	"net"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func Dial(network string, address string) (net.Conn, error) {
	addr, port, err := resolveAddress(address, "")
	if err != nil {
		return nil, err
	}

	sock, err := createSocket(network, addr)
	if err != nil {
		return nil, err
	}

	connectRes := sock.Connect(mapIpSocketAddress(addr, port))
	if connectRes.IsErr() {
		return nil, fmt.Errorf("connect: %w", mapErrorCode(connectRes.Err()))
	}

	switch s := sock.(type) {
	case *sockets.TcpSocket:
		return newTcpConn(s), nil
	case *sockets.UdpSocket:
		return newUdpConn(s), nil
	default:
		return nil, fmt.Errorf("unsupported network %q", network)
	}
}

func Listen(network string, address string) (net.Listener, error) {
	addr, port, err := resolveAddress(address, "0.0.0.0")
	if err != nil {
		return nil, err
	}

	sock, err := createTcpSocket(mapAddressFamily(addr))
	if err != nil {
		return nil, err
	}

	bindRes := sock.Bind(mapIpSocketAddress(addr, port))
	if bindRes.IsErr() {
		return nil, fmt.Errorf("bind: %w", mapErrorCode(bindRes.Err()))
	}

	listenRes := sock.Listen()
	if listenRes.IsErr() {
		return nil, fmt.Errorf("listen: %w", mapErrorCode(listenRes.Err()))
	}

	localAddrRes := sock.GetLocalAddress()
	var localAddr net.Addr
	if !localAddrRes.IsErr() {
		localAddr = mapNetAddr(localAddrRes.Ok())
	}

	return &listener{
		sock:      sock,
		stream:    listenRes.Ok(),
		localAddr: localAddr,
	}, nil
}

// wasiSocket is the subset of TcpSocket and UdpSocket shared by Dial,
// letting it connect either socket type without branching on network.
type wasiSocket interface {
	Connect(remoteAddress sockets.IpSocketAddress) witTypes.Result[witTypes.Unit, sockets.ErrorCode]
}

var (
	_ wasiSocket = (*sockets.TcpSocket)(nil)
	_ wasiSocket = (*sockets.UdpSocket)(nil)
)

func createSocket(network string, addr sockets.IpAddress) (wasiSocket, error) {
	addrFamily := mapAddressFamily(addr)
	switch network {
	case "tcp":
		return createTcpSocket(addrFamily)
	case "udp":
		return createUdpSocket(addrFamily)
	default:
		return nil, fmt.Errorf("unknown network type - %s", network)
	}
}

// newWasiSocket unwraps the create-result pattern shared by every WASI
// socket resource constructor (tcp-socket.create, udp-socket.create, ...).
func newWasiSocket[T any](res witTypes.Result[T, sockets.ErrorCode]) (T, error) {
	if res.IsErr() {
		var zero T
		return zero, fmt.Errorf("create: %w", mapErrorCode(res.Err()))
	}
	return res.Ok(), nil
}
