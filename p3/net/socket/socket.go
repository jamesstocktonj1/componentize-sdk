package socket

import (
	"fmt"
	"net"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
)

func Dial(network string, address string) (net.Conn, error) {
	addr, port, err := resolveAddress(address, "")
	if err != nil {
		return nil, err
	}

	sock, err := createTcpSocket(addr)
	if err != nil {
		return nil, err
	}

	connectRes := sock.Connect(mapIpSocketAddress(addr, port))
	if connectRes.IsErr() {
		return nil, fmt.Errorf("connect: %w", mapErrorCode(connectRes.Err()))
	}

	return newConn(sock), nil
}

func Listen(network string, address string) (net.Listener, error) {
	addr, port, err := resolveAddress(address, "0.0.0.0")
	if err != nil {
		return nil, err
	}

	sock, err := createTcpSocket(addr)
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

func createTcpSocket(addr sockets.IpAddress) (*sockets.TcpSocket, error) {
	res := sockets.TcpSocketCreate(mapAddressFamily(addr))
	if res.IsErr() {
		return nil, fmt.Errorf("create: %w", mapErrorCode(res.Err()))
	}
	return res.Ok(), nil
}
