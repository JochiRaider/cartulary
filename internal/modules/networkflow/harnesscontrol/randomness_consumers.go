package harnesscontrol

import (
	"crypto/rand"
	"errors"
	"io"
)

type controlledEntropy struct {
	registry *NetworkFlowRandomnessRegistry
	stream   string
}

func (c *Controls) TableIDEntropy() io.Reader {
	return controlledEntropy{c.Randomness, NetworkFlowRandomStreamTableID}
}
func (r controlledEntropy) Read(p []byte) (int, error) {
	if r.stream != NetworkFlowRandomStreamTableID || len(p) != 16 {
		return 0, errors.New("table entropy draw must be 16 bytes")
	}
	id, armed, err := r.registry.ConsumeNetworkFlowRandomUUID(r.stream)
	value := id[:]
	if err != nil {
		return 0, err
	}
	if !armed {
		return io.ReadFull(rand.Reader, p)
	}
	if len(value) != len(p) {
		return 0, errors.New("controlled entropy length mismatch")
	}
	return copy(p, value), nil
}
