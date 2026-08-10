package socket

import (
	"net"
	"sync"
	"time"

	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// udpConn implements net.Conn over a connected WASI UDP socket.
type udpConn struct {
	sock      *sockets.UdpSocket
	closeOnce sync.Once
}

var _ net.Conn = (*udpConn)(nil)

func createUdpSocket(addrFamily sockets.IpAddressFamily) (*sockets.UdpSocket, error) {
	return newWasiSocket(sockets.UdpSocketCreate(addrFamily))
}

func newUdpConn(sock *sockets.UdpSocket) *udpConn {
	return &udpConn{sock: sock}
}

func (c *udpConn) Read(b []byte) (int, error) {
	res := c.sock.Receive()
	if res.IsErr() {
		return 0, mapErrorCode(res.Err())
	}
	n := copy(b, res.Ok().F0)
	return n, nil
}

func (c *udpConn) Write(b []byte) (int, error) {
	res := c.sock.Send(b, witTypes.None[sockets.IpSocketAddress]())
	if res.IsErr() {
		return 0, mapErrorCode(res.Err())
	}
	return len(b), nil
}

func (c *udpConn) Close() error {
	c.closeOnce.Do(func() {
		c.sock.Drop()
	})
	return nil
}

func (c *udpConn) LocalAddr() net.Addr {
	res := c.sock.GetLocalAddress()
	if res.IsErr() {
		return nil
	}
	return mapUDPAddr(res.Ok())
}

func (c *udpConn) RemoteAddr() net.Addr {
	res := c.sock.GetRemoteAddress()
	if res.IsErr() {
		return nil
	}
	return mapUDPAddr(res.Ok())
}

func (c *udpConn) SetDeadline(t time.Time) error      { return nil }
func (c *udpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *udpConn) SetWriteDeadline(t time.Time) error { return nil }
