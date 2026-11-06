// Package portabilitysupport captures small source-port exports for focused assertions.
// Production and bulk tests use sealed, operation-owned member stores.
package portabilitysupport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type File struct {
	Path    string
	Payload []byte
}

type capture struct {
	files []File
	bytes int64
	paths map[string]bool
}

func Export(port sourceport.Port, ctx context.Context, input sourceport.ExportContext) ([]File, error) {
	destination := &capture{paths: map[string]bool{}}
	input.Members = destination
	if err := port.Export(ctx, input); err != nil {
		return nil, err
	}
	return destination.files, nil
}

func (c *capture) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	if err := ctx.Err(); err != nil {
		return incidentportability.Member{}, err
	}
	if c.paths[path] {
		return incidentportability.Member{}, fmt.Errorf("duplicate test member %q", path)
	}
	const maximum = 16 << 20
	payload, err := io.ReadAll(io.LimitReader(source, maximum-c.bytes+1))
	if err != nil {
		return incidentportability.Member{}, err
	}
	if int64(len(payload))+c.bytes > maximum {
		return incidentportability.Member{}, fmt.Errorf("focused export capture exceeds %d bytes", maximum)
	}
	c.bytes += int64(len(payload))
	c.paths[path] = true
	c.files = append(c.files, File{Path: path, Payload: payload})
	digest := sha256.Sum256(payload)
	return incidentportability.Member{Path: path, ByteSize: int64(len(payload)), SHA256: hex.EncodeToString(digest[:])}, nil
}
