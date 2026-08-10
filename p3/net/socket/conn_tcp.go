package socket

import (
	"io"
	"net"
	"sync"
	"time"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// tcpConn implements net.Conn over a WASI TCP socket.
type tcpConn struct {
	sock       *sockets.TcpSocket
	recvStream *witTypes.StreamReader[uint8]
	recvFuture *witTypes.FutureReader[witTypes.Result[witTypes.Unit, sockets.ErrorCode]]
	sendStream *witTypes.StreamWriter[uint8]
	sendFuture *witTypes.FutureReader[witTypes.Result[witTypes.Unit, sockets.ErrorCode]]
	closeOnce  sync.Once
}

var _ net.Conn = (*tcpConn)(nil)

func createTcpSocket(addrFamily sockets.IpAddressFamily) (*sockets.TcpSocket, error) {
	return newWasiSocket(sockets.TcpSocketCreate(addrFamily))
}

func newTcpConn(sock *sockets.TcpSocket) *tcpConn {
	recvStream, recvFuture := sock.Receive()
	sendStream, sendReader := sockets.MakeStreamU8()
	sendFuture := sock.Send(sendReader)
	return &tcpConn{
		sock:       sock,
		recvStream: recvStream,
		recvFuture: recvFuture,
		sendStream: sendStream,
		sendFuture: sendFuture,
	}
}

func (c *tcpConn) Read(b []byte) (int, error) {
	n := int(c.recvStream.Read(b))
	if c.recvStream.WriterDropped() {
		return n, io.EOF
	}
	return n, nil
}

func (c *tcpConn) Write(b []byte) (int, error) {
	n := int(c.sendStream.WriteAll(b))
	if c.sendStream.ReaderDropped() {
		return n, io.ErrClosedPipe
	}
	return n, nil
}

func (c *tcpConn) Close() error {
	c.closeOnce.Do(func() {
		c.sendStream.Drop()
		c.sendFuture.Drop()
		c.recvStream.Drop()
		c.recvFuture.Drop()
		c.sock.Drop()
	})
	return nil
}

func (c *tcpConn) LocalAddr() net.Addr {
	res := c.sock.GetLocalAddress()
	if res.IsErr() {
		return nil
	}
	return mapNetAddr(res.Ok())
}

func (c *tcpConn) RemoteAddr() net.Addr {
	res := c.sock.GetRemoteAddress()
	if res.IsErr() {
		return nil
	}
	return mapNetAddr(res.Ok())
}

func (c *tcpConn) SetDeadline(t time.Time) error      { return nil }
func (c *tcpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *tcpConn) SetWriteDeadline(t time.Time) error { return nil }
