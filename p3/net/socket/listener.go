package socket

import (
	"net"
	"sync"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// listener implements net.Listener over a WASI TCP socket in listening state.
type listener struct {
	sock      *sockets.TcpSocket
	stream    *witTypes.StreamReader[*sockets.TcpSocket]
	localAddr net.Addr
	closeOnce sync.Once
}

var _ net.Listener = (*listener)(nil)

func (l *listener) Accept() (net.Conn, error) {
	buf := make([]*sockets.TcpSocket, 1)
	n := l.stream.Read(buf)
	if n == 0 {
		return nil, net.ErrClosed
	}
	return newConn(buf[0]), nil
}

func (l *listener) Close() error {
	l.closeOnce.Do(func() {
		l.stream.Drop()
		l.sock.Drop()
	})
	return nil
}

func (l *listener) Addr() net.Addr {
	return l.localAddr
}
