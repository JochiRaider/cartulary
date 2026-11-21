package incidentbundles

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// errBundleMemberRetired is an internal failure: a participant tried to read a
// bundle member after its preparation returned. Core 01 REQ-01-641 forbids
// reopening a bundle member in steps 7 through 9.
var errBundleMemberRetired = errors.New("incident bundle member read after preparation")

// retiringMembers is the participant's view of its declared bundle members.
// The orchestrator retires it as soon as the participant's preparation
// returns, so the final transaction cannot reopen a bundle member.
type retiringMembers struct {
	source  incidentportability.MemberReader
	retired atomic.Bool
}

func newRetiringMembers(source incidentportability.MemberReader) *retiringMembers {
	return &retiringMembers{source: source}
}

func (m *retiringMembers) retire() { m.retired.Store(true) }

func (m *retiringMembers) Members() []incidentportability.Member { return m.source.Members() }

func (m *retiringMembers) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if m.retired.Load() {
		return nil, incidentportability.Member{}, errBundleMemberRetired
	}
	return m.source.OpenMember(ctx, path)
}

func (m *retiringMembers) VerifiesSealedContent() bool {
	verifier, ok := m.source.(incidentportability.SealVerifiedReader)
	return ok && verifier.VerifiesSealedContent()
}
