package socket

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	wasiStreams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_0_2_0_streams"
	wasiTcp "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_sockets_tcp"
)

// wasiConn implements net.Conn over WASI TCP streams.
type wasiConn struct {
	socket    *wasiTcp.TcpSocket
	reader    *wasiTcp.InputStream
	writer    *wasiTcp.OutputStream
	closeOnce sync.Once
}

var _ net.Conn = (*wasiConn)(nil)

func (c *wasiConn) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	res := c.reader.BlockingRead(uint64(len(b)))
	if res.IsErr() {
		streamErr := res.Err()
		if streamErr.Tag() == wasiStreams.StreamErrorClosed {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("read error: %v", streamErr.Tag())
	}
	data := res.Ok()
	return copy(b, data), nil
}

func (c *wasiConn) Write(b []byte) (int, error) {
	res := c.writer.BlockingWriteAndFlush(b)
	if res.IsErr() {
		return 0, fmt.Errorf("write error: %v", res.Err().Tag())
	}
	return len(b), nil
}

func (c *wasiConn) Close() error {
	c.closeOnce.Do(func() {
		c.reader.Drop()
		c.writer.Drop()
		c.socket.Drop()
	})
	return nil
}

func (c *wasiConn) LocalAddr() net.Addr {
	res := c.socket.LocalAddress()
	if res.IsErr() {
		return &net.TCPAddr{}
	}
	return ipSocketAddressToNetAddr(res.Ok())
}

func (c *wasiConn) RemoteAddr() net.Addr {
	res := c.socket.RemoteAddress()
	if res.IsErr() {
		return &net.TCPAddr{}
	}
	return ipSocketAddressToNetAddr(res.Ok())
}

func (c *wasiConn) SetDeadline(_ time.Time) error      { return nil }
func (c *wasiConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *wasiConn) SetWriteDeadline(_ time.Time) error { return nil }
