package incidentbundles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// capturingPortabilityParticipant keeps the bundle-member reader it was given.
type capturingPortabilityParticipant struct {
	*portabilityParticipantFake
	members incidentportability.MemberReader
}

func (p *capturingPortabilityParticipant) PrepareImport(ctx context.Context, invocation ImportInvocation, scope *PortabilityStagedOutputScope) (ImportPreparation, error) {
	p.members = invocation.Members
	return p.portabilityParticipantFake.PrepareImport(ctx, invocation, scope)
}

// checkParticipantMemberRetirement: a participant reads its declared bundle
// members only while it prepares. Once preparation returns, the reader is
// retired, so the final transaction cannot reopen a bundle member (Core 01
// REQ-01-641).
func checkParticipantMemberRetirement(t *testing.T) {
	policy := deadlineTestPolicy()
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	members := newArchiveTestMembers(t, Limits{})
	exporter, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true}, Participants: []ExtensionParticipant{&portabilityParticipantFake{id: policy.ParticipantID}}, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{}, ParticipantTimeout: time.Minute, FatalSink: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	if err := exporter.Export(t.Context(), portabilityQueryFake{}, incidentID, members, newActorReferenceSet()); err != nil {
		t.Fatal(err)
	}
	members.Seal()
	participant := &capturingPortabilityParticipant{portabilityParticipantFake: &portabilityParticipantFake{id: policy.ParticipantID}}
	orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true}, Participants: []ExtensionParticipant{participant}, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{}, ParticipantTimeout: time.Minute, FatalSink: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := orchestrator.PrepareImport(t.Context(), PortabilityImportContext{OperationID: "operation", IncidentID: incidentID, ActorUserID: uuid.MustParse("33333333-3333-4333-8333-333333333333")}, members)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = prepared.Abandon(context.WithoutCancel(t.Context())) }()
	if participant.members == nil || participant.readBytes == 0 || len(participant.members.Members()) == 0 {
		t.Fatal("participant did not read its declared members during preparation")
	}
	path := participant.members.Members()[0].Path
	if reader, _, err := participant.members.OpenMember(t.Context(), path); !errors.Is(err, errBundleMemberRetired) {
		if reader != nil {
			_ = reader.Close()
		}
		t.Fatalf("bundle member reader survived preparation: %v", err)
	}
}
