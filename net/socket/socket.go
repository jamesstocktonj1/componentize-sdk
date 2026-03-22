package socket

import (
	"fmt"
	"net"
	"strconv"

	instanceNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_instance_network"
	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
	wasiTcpCreate "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp_create_socket"
)

func Dial(network string, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unsupported network %q: only tcp is supported", network)
	}
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
		return nil, mapErrorCode(socketRes.Err())
	}
	sock := socketRes.Ok()

	connectRes := sock.StartConnect(n, remoteAddr)
	if connectRes.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(connectRes.Err())
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
			return nil, mapErrorCode(code)
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
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unsupported network %q: only tcp is supported", network)
	}
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", address, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	if host == "" {
		// Default to IPv4 wildcard; WASI does not support dual-stack sockets.
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
		return nil, mapErrorCode(socketRes.Err())
	}
	sock := socketRes.Ok()

	bindRes := sock.StartBind(n, localAddr)
	if bindRes.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(bindRes.Err())
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
			return nil, mapErrorCode(code)
		}
		break
	}

	listenRes := sock.StartListen()
	if listenRes.IsErr() {
		sock.Drop()
		return nil, mapErrorCode(listenRes.Err())
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
		addr:   ipSocketAddressToNetAddr(addrRes.Ok()),
	}, nil
}
