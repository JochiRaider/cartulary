package networkflow

import (
	"crypto/rand"
	"io"
	"sync"
)

// Each purpose has its own instance. A draw is indivisible even when an injected
// reader returns short reads or callers allocate concurrently.
type entropyReader struct {
	mu     sync.Mutex
	source io.Reader
}

func newEntropyReader(source io.Reader) io.Reader {
	if source == nil {
		source = rand.Reader
	}
	return &entropyReader{source: source}
}
func (r *entropyReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return io.ReadFull(r.source, p)
}
