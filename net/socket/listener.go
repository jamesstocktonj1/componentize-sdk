package socket

import (
	"net"

	wasiNetwork "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_network"
	wasiTcp "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp"
)

// wasiListener implements net.Listener over a WASI TCP socket.
type wasiListener struct {
	socket *wasiTcp.TcpSocket
	addr   net.Addr
}

var _ net.Listener = (*wasiListener)(nil)

func (l *wasiListener) Accept() (net.Conn, error) {
	for {
		res := l.socket.Accept()
		if res.IsErr() {
			code := res.Err()
			if code == wasiNetwork.ErrorCodeWouldBlock {
				pollable := l.socket.Subscribe()
				pollable.Block()
				pollable.Drop()
				continue
			}
			return nil, mapErrorCode(code)
		}
		tuple := res.Ok()
		return &wasiConn{
			socket: tuple.F0,
			reader: tuple.F1,
			writer: tuple.F2,
		}, nil
	}
}

func (l *wasiListener) Close() error {
	l.socket.Drop()
	return nil
}

func (l *wasiListener) Addr() net.Addr {
	return l.addr
}
