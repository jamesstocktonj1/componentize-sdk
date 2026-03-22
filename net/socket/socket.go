package socket

import (
	"fmt"
	"net"
	"strconv"

	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
	instanceNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_instance_network"
	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
	wasiTcp "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp"
	wasiTcpCreate "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp_create_socket"
)

func Dial(network string, address string) (net.Conn, error) {
	n, remoteAddr, err := resolveNetworkAddress(network, address, "")
	if err != nil {
		return nil, err
	}

	sock, err := createTcpSocket(remoteAddr)
	if err != nil {
		return nil, err
	}

	if res := sock.StartConnect(n, remoteAddr); res.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(res.Err())
	}

	for {
		pollable.BlockAndDrop(sock.Subscribe())

		res := sock.FinishConnect()
		if res.IsErr() {
			code := res.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, mapErrorCode(code)
		}
		streams := res.Ok()
		return &wasiConn{socket: sock, reader: streams.F0, writer: streams.F1}, nil
	}
}

func Listen(network string, address string) (net.Listener, error) {
	// Default to IPv4 wildcard; WASI does not support dual-stack sockets.
	n, localAddr, err := resolveNetworkAddress(network, address, "0.0.0.0")
	if err != nil {
		return nil, err
	}

	sock, err := createTcpSocket(localAddr)
	if err != nil {
		return nil, err
	}

	if res := sock.StartBind(n, localAddr); res.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(res.Err())
	}

	for {
		pollable.BlockAndDrop(sock.Subscribe())

		res := sock.FinishBind()
		if res.IsErr() {
			code := res.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, mapErrorCode(code)
		}
		break
	}

	if res := sock.StartListen(); res.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(res.Err())
	}

	for {
		pollable.BlockAndDrop(sock.Subscribe())

		res := sock.FinishListen()
		if res.IsErr() {
			code := res.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				continue
			}
			sock.Drop()
			return nil, mapErrorCode(code)
		}
		break
	}

	addrRes := sock.LocalAddress()
	if addrRes.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(addrRes.Err())
	}

	return &wasiListener{
		socket: sock,
		addr:   mapIpAddress(addrRes.Ok()),
	}, nil
}

func resolveNetworkAddress(network, address string, defaultHost string) (*wasiNetwork.Network, wasiNetwork.IpSocketAddress, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, wasiNetwork.IpSocketAddress{}, fmt.Errorf("unsupported network %q: only tcp is supported", network)
	}
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, wasiNetwork.IpSocketAddress{}, fmt.Errorf("invalid address %q: %w", address, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, wasiNetwork.IpSocketAddress{}, fmt.Errorf("invalid port %q: %w", portStr, err)
	}
	if host == "" {
		host = defaultHost
	}
	n := instanceNetwork.InstanceNetwork()
	addr, err := resolveAddress(n, host, uint16(port))
	if err != nil {
		return nil, wasiNetwork.IpSocketAddress{}, err
	}
	return n, addr, nil
}

func createTcpSocket(addr wasiNetwork.IpSocketAddress) (*wasiTcp.TcpSocket, error) {
	res := wasiTcpCreate.CreateTcpSocket(mapAddressFamily(addr))
	if res.IsErr() {
		return nil, mapErrorCode(res.Err())
	}
	return res.Ok(), nil
}
