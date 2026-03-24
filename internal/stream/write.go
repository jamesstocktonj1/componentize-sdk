package stream

import (
	"fmt"
	"io"

	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_0_2_0_streams"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
)

// WriteStream writes p to stream in capacity-limited chunks, blocking as needed,
// then flushes and waits for the flush to complete.
func WriteStream(s *streams.OutputStream, p []byte) (int, error) {
	written := 0
	for written < len(p) {
		checkRes := s.CheckWrite()
		if checkRes.IsErr() {
			if checkRes.Err().Tag() == streams.StreamErrorClosed {
				return written, io.EOF
			}
			return written, fmt.Errorf("failed to check write capacity: %v", checkRes.Err())
		}
		capacity := checkRes.Ok()
		if capacity == 0 {
			pollable.AwaitAndDrop(s.Subscribe())
			continue
		}
		chunk := p[written:]
		if uint64(len(chunk)) > capacity {
			chunk = chunk[:capacity]
		}
		writeRes := s.Write(chunk)
		if writeRes.IsErr() {
			if writeRes.Err().Tag() == streams.StreamErrorClosed {
				return written, io.EOF
			}
			return written, fmt.Errorf("failed to write to stream: %v", writeRes.Err())
		}
		written += len(chunk)
	}

	if flushRes := s.Flush(); flushRes.IsErr() {
		if flushRes.Err().Tag() == streams.StreamErrorClosed {
			return written, io.EOF
		}
		return written, fmt.Errorf("failed to flush stream: %v", flushRes.Err())
	}
	pollable.AwaitAndDrop(s.Subscribe())
	return written, nil
}
