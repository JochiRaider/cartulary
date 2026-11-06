package portabilitysupport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"io"
	"sort"
)

// MapBundle is a small-fixture helper. Production uses sealed member readers.
type MapBundle map[string][]byte

func (b MapBundle) Members() []incidentportability.Member {
	members := make([]incidentportability.Member, 0, len(b))
	for _, path := range b.Paths() {
		raw := b[path]
		digest := sha256.Sum256(raw)
		members = append(members, incidentportability.Member{Path: path, ByteSize: int64(len(raw)), SHA256: hex.EncodeToString(digest[:])})
	}
	return members
}
func (b MapBundle) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, incidentportability.Member{}, err
	}
	raw, ok := b[path]
	if !ok {
		return nil, incidentportability.Member{}, &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
	}
	digest := sha256.Sum256(raw)
	return io.NopCloser(bytes.NewReader(raw)), incidentportability.Member{Path: path, ByteSize: int64(len(raw)), SHA256: hex.EncodeToString(digest[:])}, nil
}

func (b MapBundle) File(path string) ([]byte, bool) {
	payload, ok := b[path]
	return payload, ok
}

func (b MapBundle) Paths() []string {
	paths := make([]string, 0, len(b))
	for path := range b {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
