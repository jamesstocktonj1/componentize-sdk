package socket

import (
	"io"
	"net"
	"sync"
	"time"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// conn implements net.Conn over a WASI TCP socket.
type conn struct {
	sock       *sockets.TcpSocket
	recvStream *witTypes.StreamReader[uint8]
	recvFuture *witTypes.FutureReader[witTypes.Result[witTypes.Unit, sockets.ErrorCode]]
	sendStream *witTypes.StreamWriter[uint8]
	sendFuture *witTypes.FutureReader[witTypes.Result[witTypes.Unit, sockets.ErrorCode]]
	closeOnce  sync.Once
}

var _ net.Conn = (*conn)(nil)

func newConn(sock *sockets.TcpSocket) *conn {
	recvStream, recvFuture := sock.Receive()
	sendStream, sendReader := sockets.MakeStreamU8()
	sendFuture := sock.Send(sendReader)
	return &conn{
		sock:       sock,
		recvStream: recvStream,
		recvFuture: recvFuture,
		sendStream: sendStream,
		sendFuture: sendFuture,
	}
}

func (c *conn) Read(b []byte) (int, error) {
	n := int(c.recvStream.Read(b))
	if c.recvStream.WriterDropped() {
		return n, io.EOF
	}
	return n, nil
}

func (c *conn) Write(b []byte) (int, error) {
	n := int(c.sendStream.WriteAll(b))
	if c.sendStream.ReaderDropped() {
		return n, io.ErrClosedPipe
	}
	return n, nil
}

func (c *conn) Close() error {
	c.closeOnce.Do(func() {
		c.sendStream.Drop()
		c.sendFuture.Drop()
		c.recvStream.Drop()
		c.recvFuture.Drop()
		c.sock.Drop()
	})
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	res := c.sock.GetLocalAddress()
	if res.IsErr() {
		return nil
	}
	return mapNetAddr(res.Ok())
}

func (c *conn) RemoteAddr() net.Addr {
	res := c.sock.GetRemoteAddress()
	if res.IsErr() {
		return nil
	}
	return mapNetAddr(res.Ok())
}

func (c *conn) SetDeadline(t time.Time) error      { return nil }
func (c *conn) SetReadDeadline(t time.Time) error  { return nil }
func (c *conn) SetWriteDeadline(t time.Time) error { return nil }

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
