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
func (c *Controls) CursorNonceEntropy() io.Reader {
	return controlledEntropy{c.Randomness, NetworkFlowRandomStreamCursorNonce}
}
func (r controlledEntropy) Read(p []byte) (int, error) {
	var value []byte
	var armed bool
	var err error
	if r.stream == NetworkFlowRandomStreamTableID {
		if len(p) != 16 {
			return 0, errors.New("table entropy draw must be 16 bytes")
		}
		id, ok, consumeErr := r.registry.ConsumeNetworkFlowRandomUUID(r.stream)
		value, armed, err = id[:], ok, consumeErr
	} else {
		if len(p) != 12 {
			return 0, errors.New("cursor entropy draw must be 12 bytes")
		}
		value, armed, err = r.registry.ConsumeNetworkFlowRandomHexBytes(r.stream)
	}
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
