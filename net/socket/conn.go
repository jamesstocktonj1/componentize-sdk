package socket

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
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
	waitable := c.reader.Subscribe()
	defer waitable.Drop()
	if err := pollable.Await(waitable); err != nil {
		return 0, err
	}
	res := c.reader.Read(uint64(len(b)))
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
	written := 0
	for written < len(b) {
		checkRes := c.writer.CheckWrite()
		if checkRes.IsErr() {
			return written, fmt.Errorf("write check error: %v", checkRes.Err().Tag())
		}
		capacity := checkRes.Ok()
		if capacity == 0 {
			waitable := c.writer.Subscribe()
			if err := pollable.Await(waitable); err != nil {
				waitable.Drop()
				return written, err
			}
			waitable.Drop()
			continue
		}
		chunk := b[written:]
		if uint64(len(chunk)) > capacity {
			chunk = chunk[:capacity]
		}
		res := c.writer.Write(chunk)
		if res.IsErr() {
			return written, fmt.Errorf("write error: %v", res.Err().Tag())
		}
		written += len(chunk)
	}

	// Flush and wait for completion.
	if flushRes := c.writer.Flush(); flushRes.IsErr() {
		return written, fmt.Errorf("flush error: %v", flushRes.Err().Tag())
	}
	waitable := c.writer.Subscribe()
	defer waitable.Drop()
	if err := pollable.Await(waitable); err != nil {
		return written, err
	}
	return written, nil
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
	return mapIpAddress(res.Ok())
}

func (c *wasiConn) RemoteAddr() net.Addr {
	res := c.socket.RemoteAddress()
	if res.IsErr() {
		return &net.TCPAddr{}
	}
	return mapIpAddress(res.Ok())
}

func (c *wasiConn) SetDeadline(_ time.Time) error      { return nil }
func (c *wasiConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *wasiConn) SetWriteDeadline(_ time.Time) error { return nil }
