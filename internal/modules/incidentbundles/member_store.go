package incidentbundles

import (
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
)

type memberStaging interface {
	AllocateStream(context.Context, string, io.Reader, int64) (stagedobjects.SealedReference, error)
	OpenSealed(context.Context, string, stagedobjects.SealedReference) (io.ReadCloser, error)
}

// memberStore retains only sealed references. The existing staged-object scope
// owns all bytes, leases, cleanup and later transfer of publication ownership.
type memberBudget struct{ bytes, members int64 }

type memberStore struct {
	operationID    string
	staging        memberStaging
	limits         Limits
	members        map[string]stagedobjects.SealedReference
	total          int64
	sealed         bool
	pathPrefix     string
	sharedBudget   *memberBudget
	remainingItems *int
	// reads observes member opens by store class and phase for the transfer
	// evidence recorder. Production composition installs none.
	reads      MemberReadObserver
	class      string
	finalPhase atomic.Bool
}

// MemberReadObserver receives one event per member open: the import operation,
// the store ("bundle", "rewritten" or "prepared", or "payload" for an
// admitted opaque payload streamed from its sealed bundle member), the phase
// ("preparation", or "final_transaction" once the cross-owner protocol has
// started) and the opened member's path.
type MemberReadObserver func(operationID, store, phase, path string)

func (s *memberStore) observeReads(class string, observer MemberReadObserver) {
	s.class, s.reads = class, observer
}

// enterFinalTransaction marks the start of the cross-owner final-commit
// protocol; later opens are reported in the final_transaction phase.
func (s *memberStore) enterFinalTransaction() { s.finalPhase.Store(true) }

// VerifiesSealedContent reports that opened streams are verified against the
// immutable staged seal, so readers need not hash them again.
func (s *memberStore) VerifiesSealedContent() bool {
	verifier, ok := s.staging.(interface{ VerifiesSealedContent() bool })
	return ok && verifier.VerifiesSealedContent()
}

func newMemberStore(operationID string, staging memberStaging, limits Limits) (*memberStore, error) {
	if operationID == "" || staging == nil {
		return nil, stagedobjects.ErrScopeDenied
	}
	return &memberStore{operationID: operationID, staging: staging, limits: limits, members: map[string]stagedobjects.SealedReference{}}, nil
}
func (s *memberStore) maximumBytes() int64 {
	if s.limits.IncidentBundles.MaxExtractedBytes > 0 {
		return s.limits.IncidentBundles.MaxExtractedBytes
	}
	return defaultIncidentBundleMaxExtractedBytes
}
func (s *memberStore) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	if s.sealed {
		return incidentportability.Member{}, stagedobjects.ErrScopeDenied
	}
	if !safeBundlePath(path) {
		return incidentportability.Member{}, &verificationError{ReasonCode: "invalid_member_path"}
	}
	if s.pathPrefix != "" && !strings.HasPrefix(path, s.pathPrefix) {
		return incidentportability.Member{}, stagedobjects.ErrScopeDenied
	}
	if _, exists := s.members[path]; exists {
		return incidentportability.Member{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if err := checkMemberCount(len(s.members)+1, s.limits); err != nil {
		return incidentportability.Member{}, err
	}
	if s.remainingItems != nil && *s.remainingItems <= 0 {
		return incidentportability.Member{}, ErrPortabilityLimit
	}
	remaining := s.maximumBytes() - s.total
	if s.sharedBudget != nil {
		if s.sharedBudget.members <= 0 {
			return incidentportability.Member{}, &verificationError{ReasonCode: "archive_member_count_exceeded"}
		}
		remaining = min(remaining, s.sharedBudget.bytes)
	}
	if remaining < 0 {
		return incidentportability.Member{}, &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
	}
	// A zero-byte member is legal even at the exact cumulative ceiling. A probe
	// still runs through staging, but no nonempty object can enter the catalog.
	ref, err := s.staging.AllocateStream(ctx, s.operationID, source, max(remaining, 1))
	if err != nil {
		var failure *stagedobjects.Failure
		if errors.As(err, &failure) && failure.SafeCode == "member_bytes_exceeded" {
			return incidentportability.Member{}, &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		return incidentportability.Member{}, err
	}
	if ref.ByteSize > remaining {
		return incidentportability.Member{}, &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
	}
	s.members[path] = ref
	s.total += ref.ByteSize
	if s.remainingItems != nil {
		(*s.remainingItems)--
	}
	if s.sharedBudget != nil {
		s.sharedBudget.bytes -= ref.ByteSize
		s.sharedBudget.members--
	}
	return incidentportability.Member{Path: path, ByteSize: ref.ByteSize, SHA256: ref.SHA256}, nil
}
func (s *memberStore) Members() []incidentportability.Member {
	result := make([]incidentportability.Member, 0, len(s.members))
	for path, ref := range s.members {
		result = append(result, incidentportability.Member{Path: path, ByteSize: ref.ByteSize, SHA256: ref.SHA256})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
func (s *memberStore) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	ref, ok := s.members[path]
	if !ok {
		return nil, incidentportability.Member{}, &verificationError{ReasonCode: "missing_required_file"}
	}
	if s.reads != nil {
		phase := "preparation"
		if s.finalPhase.Load() {
			phase = "final_transaction"
		}
		s.reads(s.operationID, s.class, phase, path)
	}
	source, err := s.staging.OpenSealed(ctx, s.operationID, ref)
	return source, incidentportability.Member{Path: path, ByteSize: ref.ByteSize, SHA256: ref.SHA256}, err
}
func (s *memberStore) Seal() { s.sealed = true }

// member returns one staged member's sealed descriptor.
func (s *memberStore) member(path string) (incidentportability.Member, bool) {
	ref, ok := s.members[path]
	if !ok {
		return incidentportability.Member{}, false
	}
	return incidentportability.Member{Path: path, ByteSize: ref.ByteSize, SHA256: ref.SHA256}, true
}

// openPayload streams one admitted opaque payload. Opens are observed as
// "payload" reads, never as row reads of the bundle (Core 01 REQ-01-641).
func (s *memberStore) openPayload(ctx context.Context, expected incidentportability.Member) (io.ReadCloser, error) {
	ref, ok := s.members[expected.Path]
	if !ok || ref.ByteSize != expected.ByteSize || ref.SHA256 != expected.SHA256 {
		return nil, incidentportability.ErrPayloadNotAdmitted
	}
	if s.reads != nil {
		phase := "preparation"
		if s.finalPhase.Load() {
			phase = "final_transaction"
		}
		s.reads(s.operationID, "payload", phase, expected.Path)
	}
	return s.staging.OpenSealed(ctx, s.operationID, ref)
}

// Only control catalogs may be materialized. Bulk callers receive readers.
func readControlMember(ctx context.Context, members incidentportability.MemberReader, path string, maximum int64) ([]byte, error) {
	source, member, err := members.OpenMember(ctx, path)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	if maximum < 1 || maximum > PortabilityParticipantByteLimit || member.ByteSize > maximum {
		return nil, &verificationError{ReasonCode: "malformed_manifest"}
	}
	raw, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) != member.ByteSize {
		return nil, &verificationError{ReasonCode: "checksum_mismatch"}
	}
	return raw, nil
}

var _ incidentportability.MemberReader = (*memberStore)(nil)
var _ incidentportability.MemberWriter = (*memberStore)(nil)
