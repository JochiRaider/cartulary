package incidentbundles

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// errPayloadAdmissionClosed is an internal failure: preparation admits
// payloads, and admission closes before the final transaction begins.
var errPayloadAdmissionClosed = errors.New("incident bundle payload admission closed")

// payloadMembers is the operation's opaque-payload view of the sealed bundle
// (Core 01 REQ-01-641). Preparation admits a member whose digest it verified
// and copies nothing; the final transaction then streams only admitted
// members, each verified against its seal and hashed once while read. It
// stays open after the row reader retires and opens nothing else.
type payloadMembers struct {
	members  *memberStore
	mu       sync.Mutex
	admitted map[string]incidentportability.Member
	closed   bool
}

func newPayloadMembers(members *memberStore) *payloadMembers {
	return &payloadMembers{members: members, admitted: map[string]incidentportability.Member{}}
}

func (p *payloadMembers) AdmitPayload(_ context.Context, path string) (incidentportability.Member, error) {
	if p == nil || p.members == nil {
		return incidentportability.Member{}, incidentportability.ErrPayloadNotAdmitted
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return incidentportability.Member{}, errPayloadAdmissionClosed
	}
	member, ok := p.members.member(path)
	if !ok {
		return incidentportability.Member{}, &verificationError{ReasonCode: "missing_required_file"}
	}
	p.admitted[path] = member
	return member, nil
}

func (p *payloadMembers) OpenPayload(ctx context.Context, member incidentportability.Member) (io.ReadCloser, error) {
	if p == nil || p.members == nil {
		return nil, incidentportability.ErrPayloadNotAdmitted
	}
	p.mu.Lock()
	admitted, ok := p.admitted[member.Path]
	p.mu.Unlock()
	if !ok || admitted != member {
		return nil, incidentportability.ErrPayloadNotAdmitted
	}
	return p.members.openPayload(ctx, member)
}

// closeAdmission ends preparation's admissions; admitted payloads stay
// readable.
func (p *payloadMembers) closeAdmission() {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
}

var _ incidentportability.PayloadMembers = (*payloadMembers)(nil)
