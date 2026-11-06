package incidentportability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"sort"
)

// OpenVerifiedMember binds a reusable owner preparation to exactly the sealed
// member it admitted. The stream also verifies bytes rather than trusting a
// caller-supplied catalog. Publication callers must consume through EOF.
func OpenVerifiedMember(ctx context.Context, members MemberReader, expected Member) (io.ReadCloser, error) {
	if members == nil || expected.Path == "" || expected.ByteSize < 0 || len(expected.SHA256) != 64 {
		return nil, &VerificationFailure{ReasonCode: "checksum_mismatch"}
	}
	source, actual, err := members.OpenMember(ctx, expected.Path)
	if err != nil {
		return nil, err
	}
	if source == nil || actual != expected {
		if source != nil {
			_ = source.Close()
		}
		return nil, &VerificationFailure{ReasonCode: "checksum_mismatch"}
	}
	return &verifiedMemberRead{ctx: ctx, source: source, expected: expected, digest: sha256.New()}, nil
}

type verifiedMemberRead struct {
	ctx      context.Context
	source   io.ReadCloser
	expected Member
	digest   hash.Hash
	count    int64
	terminal error
}

func (r *verifiedMemberRead) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if r.terminal != nil {
		return 0, r.terminal
	}
	n, err := r.source.Read(p)
	if n > 0 {
		r.count += int64(n)
		_, _ = r.digest.Write(p[:n])
	}
	if r.count > r.expected.ByteSize {
		r.terminal = &VerificationFailure{ReasonCode: "checksum_mismatch"}
		return 0, r.terminal
	}
	if errors.Is(err, io.EOF) {
		if r.count != r.expected.ByteSize || hex.EncodeToString(r.digest.Sum(nil)) != r.expected.SHA256 {
			err = &VerificationFailure{ReasonCode: "checksum_mismatch"}
		}
	}
	if err != nil {
		r.terminal = err
	}
	return n, err
}
func (r *verifiedMemberRead) Close() error { return r.source.Close() }

// WalkMember reopens one admitted immutable record member without materializing
// its bytes. Owners validate each decoded row and retain only their descriptors.
func WalkMember(ctx context.Context, members MemberReader, expected Member, visit func(map[string]any) error) error {
	source, err := OpenVerifiedMember(ctx, members, expected)
	if err != nil {
		return err
	}
	err = WalkNDJSON(ctx, source, expected.Path, MaxNDJSONLogicalLineBytes, visit)
	return errors.Join(err, source.Close())
}

// BoundMembers keeps only immutable descriptors for an owner's admitted paths.
// A reopened member must still match the original descriptor and exact bytes.
type BoundMembers struct {
	source  MemberReader
	members map[string]Member
}

func BindMembers(source MemberReader, paths []string) (*BoundMembers, error) {
	if source == nil {
		return nil, &VerificationFailure{ReasonCode: "missing_required_file"}
	}
	catalog := map[string]Member{}
	for _, member := range source.Members() {
		if _, duplicate := catalog[member.Path]; duplicate {
			return nil, &VerificationFailure{ReasonCode: "malformed_manifest"}
		}
		catalog[member.Path] = member
	}
	bound := &BoundMembers{source: source, members: map[string]Member{}}
	for _, path := range paths {
		member, ok := catalog[path]
		if !ok {
			return nil, &VerificationFailure{ReasonCode: "missing_required_file"}
		}
		bound.members[path] = member
	}
	return bound, nil
}
func (b *BoundMembers) Members() []Member {
	result := make([]Member, 0, len(b.members))
	for _, member := range b.members {
		result = append(result, member)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
func (b *BoundMembers) OpenMember(ctx context.Context, path string) (io.ReadCloser, Member, error) {
	member, ok := b.members[path]
	if !ok {
		return nil, Member{}, &VerificationFailure{ReasonCode: "missing_required_file"}
	}
	source, err := OpenVerifiedMember(ctx, b.source, member)
	return source, member, err
}
func WalkPath(ctx context.Context, source MemberReader, path string, visit func(map[string]any) error) error {
	bound, err := BindMembers(source, []string{path})
	if err != nil {
		return err
	}
	stream, _, err := bound.OpenMember(ctx, path)
	if err != nil {
		return err
	}
	err = WalkNDJSON(ctx, stream, path, MaxNDJSONLogicalLineBytes, visit)
	return errors.Join(err, stream.Close())
}

const MaxControlMemberBytes int64 = 64 << 20

// ReadControl is reserved for schema-bounded metadata, never bulk records.
func ReadControl(ctx context.Context, source MemberReader, path string, maximum int64) ([]byte, error) {
	if maximum < 1 || maximum > MaxControlMemberBytes {
		return nil, &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	bound, err := BindMembers(source, []string{path})
	if err != nil {
		return nil, err
	}
	member := bound.members[path]
	if member.ByteSize > maximum {
		return nil, &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	stream, _, err := bound.OpenMember(ctx, path)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(stream, maximum+1))
	err = errors.Join(err, stream.Close())
	if err == nil && int64(len(raw)) != member.ByteSize {
		err = &VerificationFailure{ReasonCode: "checksum_mismatch"}
	}
	return raw, err
}
