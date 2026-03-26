package pollable

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockReadable struct {
	sync.Mutex
	hasDropped bool
	isReady    bool
}

var _ Readable = (*mockReadable)(nil)

func (r *mockReadable) SetReady() {
	r.Lock()
	defer r.Unlock()

	r.isReady = true
}

func (r *mockReadable) Ready() bool {
	return r.isReady
}

func (r *mockReadable) Drop() {
	r.Lock()
	defer r.Unlock()

	r.hasDropped = true
}

func TestAwaitContext(t *testing.T) {
	r := &mockReadable{}

	go func() {
		time.Sleep(time.Millisecond)
		r.SetReady()
	}()

	err := AwaitContext(context.Background(), r)
	assert.NoError(t, err)
	assert.False(t, r.hasDropped)
}
