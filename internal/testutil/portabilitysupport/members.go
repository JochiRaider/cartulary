package portabilitysupport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
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

// PreparationStore is a focused-test core preparation store. Production uses
// the operation's sealed staged-object member store.
type PreparationStore struct {
	mu    sync.Mutex
	files MapBundle
}

func NewPreparationStore() *PreparationStore { return &PreparationStore{files: MapBundle{}} }

func (s *PreparationStore) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	const maximum = 64 << 20
	payload, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return incidentportability.Member{}, err
	}
	if len(payload) > maximum {
		return incidentportability.Member{}, fmt.Errorf("focused preparation member %q exceeds %d bytes", path, maximum)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.files[path]; exists {
		return incidentportability.Member{}, fmt.Errorf("duplicate preparation member %q", path)
	}
	s.files[path] = payload
	digest := sha256.Sum256(payload)
	return incidentportability.Member{Path: path, ByteSize: int64(len(payload)), SHA256: hex.EncodeToString(digest[:])}, ctx.Err()
}

func (s *PreparationStore) Members() []incidentportability.Member {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files.Members()
}

func (s *PreparationStore) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files.OpenMember(ctx, path)
}

// SequencedPreparationStore places each written member under its own
// sequence prefix, as the source adapter scopes each preparation, for tests
// that call an owner's preparation directly more than once with one store.
type SequencedPreparationStore struct {
	*PreparationStore
	next atomic.Int64
}

func NewSequencedPreparationStore() *SequencedPreparationStore {
	return &SequencedPreparationStore{PreparationStore: NewPreparationStore()}
}

func (s *SequencedPreparationStore) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	return s.PreparationStore.WriteMember(ctx, fmt.Sprintf("%d/%s", s.next.Add(1), path), source)
}

// CountingBundle counts bundle member opens after Freeze, which a test calls
// when preparation ends. Core 01 REQ-01-641 forbids any such reopen.
type CountingBundle struct {
	incidentportability.MemberReader
	frozen atomic.Bool
	late   atomic.Int64
}

func (c *CountingBundle) Freeze()     { c.frozen.Store(true) }
func (c *CountingBundle) Late() int64 { return c.late.Load() }

func (c *CountingBundle) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if c.frozen.Load() {
		c.late.Add(1)
	}
	return c.MemberReader.OpenMember(ctx, path)
}

// Tamper replaces the bytes of the first staged member whose path ends with
// suffix, for tests of seal verification. It reports whether one was found.
func (s *PreparationStore) Tamper(suffix string, payload []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, path := range s.files.Paths() {
		if strings.HasSuffix(path, suffix) {
			s.files[path] = payload
			return true
		}
	}
	return false
}

// PayloadMembers is a focused test's opaque-payload view of a bundle (Core 01
// REQ-01-641). It admits members of the bundle and opens only admitted ones,
// verifying their digests while read; Opens counts the payload opens.
type PayloadMembers struct {
	bundle   incidentportability.MemberReader
	mu       sync.Mutex
	admitted map[string]incidentportability.Member
	opens    atomic.Int64
}

func NewPayloadMembers(bundle incidentportability.MemberReader) *PayloadMembers {
	return &PayloadMembers{bundle: bundle, admitted: map[string]incidentportability.Member{}}
}

func (p *PayloadMembers) AdmitPayload(_ context.Context, path string) (incidentportability.Member, error) {
	for _, member := range p.bundle.Members() {
		if member.Path == path {
			p.mu.Lock()
			p.admitted[path] = member
			p.mu.Unlock()
			return member, nil
		}
	}
	return incidentportability.Member{}, fmt.Errorf("payload %q is not a bundle member", path)
}

func (p *PayloadMembers) OpenPayload(ctx context.Context, member incidentportability.Member) (io.ReadCloser, error) {
	p.mu.Lock()
	admitted, ok := p.admitted[member.Path]
	p.mu.Unlock()
	if !ok || admitted != member {
		return nil, incidentportability.ErrPayloadNotAdmitted
	}
	p.opens.Add(1)
	return incidentportability.OpenVerifiedMember(ctx, p.bundle, member)
}

func (p *PayloadMembers) Opens() int64 { return p.opens.Load() }
