package crawler

import (
	"bytes"
	"io"
	"sync"
)

type tailBuffer struct {
	maxBytes int
	mu       sync.Mutex
	buf      []byte
}

func newTailBuffer(maxBytes int) *tailBuffer {
	if maxBytes <= 0 {
		maxBytes = 64 << 10
	}
	return &tailBuffer{maxBytes: maxBytes, buf: make([]byte, 0, maxBytes)}
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(p) >= t.maxBytes {
		t.buf = append(t.buf[:0], p[len(p)-t.maxBytes:]...)
		return len(p), nil
	}

	t.buf = append(t.buf, p...)
	if len(t.buf) > t.maxBytes {
		t.buf = append(t.buf[:0], t.buf[len(t.buf)-t.maxBytes:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(bytes.TrimSpace(t.buf))
}

var _ io.Writer = (*tailBuffer)(nil)
