package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
)

func TestIncidentBundlePortabilityStateAndClaimMatrix_Integration(t *testing.T) {
	incidentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	base := ExtensionPolicy{
		ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v1",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v1",
		MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit,
		AuthoritativeFamilyIDs: []string{"profile.state"},
	}
	tests := []struct {
		name        string
		policy      ExtensionPolicy
		present     bool
		wantPayload bool
		wantError   error
		wantCalls   int
	}{
		{name: "no-authoritative absent", policy: withPortabilityMode(base, PortabilityModeNoAuthoritativeState), present: false},
		{name: "no-authoritative integrity", policy: withPortabilityMode(base, PortabilityModeNoAuthoritativeState), present: true, wantError: ErrPortabilityResult},
		{name: "participant absent claimed", policy: base, present: false},
		{name: "participant absent unclaimed", policy: withClaim(base, ClaimStateUnclaimed), present: false},
		{name: "participant present claimed", policy: base, present: true, wantPayload: true, wantCalls: 1},
		{name: "participant present unclaimed", policy: withClaim(base, ClaimStateUnclaimed), present: true, wantError: ErrPortabilityUnavailable},
		{name: "participant present unclaimable", policy: withClaim(base, ClaimStateRecognizedUnclaimable), present: true, wantError: ErrPortabilityUnavailable},
		{name: "blocked absent", policy: withPortabilityMode(base, PortabilityModeBlockedWhenPresent), present: false},
		{name: "blocked present", policy: withPortabilityMode(base, PortabilityModeBlockedWhenPresent), present: true, wantError: ErrPortabilityBlocked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			participant := &portabilityParticipantFake{id: base.ParticipantID}
			participants := []ExtensionParticipant(nil)
			if test.policy.Mode == PortabilityModeParticipant && test.policy.ClaimState == ClaimStateClaimed {
				participants = append(participants, participant)
			}
			orchestrator, err := NewPortabilityOrchestrator(
				[]ExtensionPolicy{test.policy},
				portabilityPresenceFake{present: test.present},
				participants,
				&portabilityAllocatorFake{},
			)
			if err != nil {
				t.Fatal(err)
			}
			payloads, err := orchestrator.Export(context.Background(), incidentID)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v; want %v", err, test.wantError)
			}
			if got := len(payloads) == 1; got != test.wantPayload {
				t.Fatalf("payload present = %t; want %t (%#v)", got, test.wantPayload, payloads)
			}
			if participant.exportCalls != test.wantCalls {
				t.Fatalf("participant calls = %d; want %d", participant.exportCalls, test.wantCalls)
			}
		})
	}
}

func TestIncidentBundlePortabilityImportAdmissionAndCleanup_Integration(t *testing.T) {
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	payload := ExtensionPayload{
		ProfileID: "profile", ContractMajor: 1, StateVersion: 1,
		PayloadSchemaID: "profile.payload.v1", Payload: []byte(`{"value":"portable"}`),
	}
	payload.PayloadSHA256 = digestBytes(payload.Payload)
	filePath, encoded, err := EncodeExtensionPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{filePath: encoded}
	policy := ExtensionPolicy{
		ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v1",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v1",
		MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit,
	}
	allocator := &portabilityAllocatorFake{}
	participant := &portabilityParticipantFake{id: policy.ParticipantID, allocate: true}
	orchestrator, err := NewPortabilityOrchestrator([]ExtensionPolicy{policy}, portabilityPresenceFake{}, []ExtensionParticipant{participant}, allocator)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := orchestrator.PrepareImport(context.Background(), "operation", incidentID, files)
	if err != nil {
		t.Fatal(err)
	}
	if participant.importCalls != 1 || len(prepared.Participants) != 1 || len(prepared.Transfers) != 1 {
		t.Fatalf("preparation = calls %d participants %d transfers %d", participant.importCalls, len(prepared.Participants), len(prepared.Transfers))
	}
	if refs := prepared.Transfers[0].References(); len(refs) != 1 ||
		!strings.HasPrefix(participant.stagedRefs[0], "cartulary:staged_output:") {
		t.Fatalf("staged references were not kept behind logical refs: transfer=%v participant=%v", refs, participant.stagedRefs)
	}
	input, err := prepared.Participants[0].BuildInput(context.Background(), crossownertransaction.OperationContext{})
	if err != nil || string(input.CanonicalBytes) != `{"prepared":true}` {
		t.Fatalf("prepared transaction input was not bound: %#v/%v", input, err)
	}
	if err := prepared.Abandon(context.Background()); err != nil || len(allocator.abandoned) != 1 {
		t.Fatalf("abandon = %v refs=%v", err, allocator.abandoned)
	}

	for _, test := range []struct {
		name   string
		policy ExtensionPolicy
		files  map[string][]byte
	}{
		{name: "unknown profile", policy: policy, files: map[string][]byte{"ext/extensions/unknown/payload.json": replaceProfile(encoded, "profile", "unknown")}},
		{name: "recognized unclaimable", policy: withClaim(policy, ClaimStateRecognizedUnclaimable), files: files},
		{name: "unclaimed", policy: withClaim(policy, ClaimStateUnclaimed), files: files},
		{name: "nonparticipant", policy: withPortabilityMode(policy, PortabilityModeNoAuthoritativeState), files: files},
		{name: "incompatible major", policy: policy, files: map[string][]byte{filePath: replaceMajor(encoded, 2)}},
		{name: "digest mismatch", policy: policy, files: map[string][]byte{filePath: replaceDigest(encoded, strings.Repeat("f", 64))}},
	} {
		t.Run(test.name, func(t *testing.T) {
			participants := []ExtensionParticipant(nil)
			if test.policy.Mode == PortabilityModeParticipant && test.policy.ClaimState == ClaimStateClaimed {
				participants = append(participants, &portabilityParticipantFake{id: policy.ParticipantID})
			}
			current, err := NewPortabilityOrchestrator([]ExtensionPolicy{test.policy}, portabilityPresenceFake{}, participants, &portabilityAllocatorFake{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := current.PrepareImport(context.Background(), "operation", incidentID, test.files); err == nil {
				t.Fatal("incompatible payload was accepted")
			}
		})
	}

	payload.Payload = []byte(`"` + strings.Repeat("a", PortabilityParticipantByteLimit) + `"`)
	payload.PayloadSHA256 = digestBytes(payload.Payload)
	filePath, encoded, err = EncodeExtensionPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orchestrator.PrepareImport(context.Background(), "operation", incidentID, map[string][]byte{filePath: encoded}); !errors.Is(err, ErrPortabilityLimit) {
		t.Fatalf("64 MiB plus one = %v", err)
	}
}

func TestIncidentBundleImportDescriptorIsClosed_Unit(t *testing.T) {
	descriptor := ImportTransactionDescriptor()
	if descriptor.ParticipantID != ImportTransactionParticipantID ||
		len(descriptor.SerializationKeyKinds) != 1 ||
		len(descriptor.OwnedStateFamilyIDs) != 2 ||
		len(descriptor.ContractSHA256) != 64 {
		t.Fatalf("descriptor = %#v", descriptor)
	}
}

type portabilityPresenceFake struct{ present bool }

func (p portabilityPresenceFake) AuthoritativeStatePresent(context.Context, uuid.UUID, string, []string) (bool, error) {
	return p.present, nil
}

type portabilityAllocatorFake struct {
	allocated []string
	abandoned []string
}

func (a *portabilityAllocatorFake) Allocate(_ context.Context, operationID, profileID string, _ []byte) (stagedobjects.Reference, error) {
	ref := operationID + ":" + profileID
	a.allocated = append(a.allocated, ref)
	return stagedobjects.Reference{StagingID: ref}, nil
}

func (a *portabilityAllocatorFake) Abandon(_ context.Context, reference stagedobjects.Reference) error {
	a.abandoned = append(a.abandoned, reference.StagingID)
	return nil
}

type portabilityParticipantFake struct {
	id          string
	exportCalls int
	importCalls int
	allocate    bool
	stagedRefs  []string
}

func (p *portabilityParticipantFake) ID() string { return p.id }

func (p *portabilityParticipantFake) ContractSHA256() string {
	return strings.Repeat("a", 64)
}

func (p *portabilityParticipantFake) SpecializationSchemaID() string {
	return "test.specialization.v1"
}

func (p *portabilityParticipantFake) Export(context.Context, ExportInvocation) (ExportResult, error) {
	p.exportCalls++
	return ExportResult{
		SchemaID: ExtensionExportResultSchema, Kind: "payload",
		PayloadSchemaID: "profile.payload.v1", PayloadContractMajor: 1,
		StateVersion: 1, Payload: []byte(`{"value":"portable"}`),
	}, nil
}

func (p *portabilityParticipantFake) PrepareImport(ctx context.Context, invocation ImportInvocation, scope *PortabilityStagedOutputScope) (ImportPreparation, error) {
	p.importCalls++
	if p.allocate {
		if _, err := scope.Allocate(ctx, invocation.OperationID, []byte("staged")); err != nil {
			return ImportPreparation{}, err
		}
	}
	p.stagedRefs = scope.Refs()
	input := []byte(`{"prepared":true}`)
	return ImportPreparation{
		SchemaID: ExtensionImportResultSchema, Status: "prepared",
		ParticipantInput: input, ParticipantInputSHA256: digestBytes(input),
		StagedOutputRefs: scope.Refs(), TransactionParticipant: portabilityTransactionParticipantFake{},
	}, nil
}

type portabilityTransactionParticipantFake struct{}

func (portabilityTransactionParticipantFake) ID() string { return "profile.import_v1" }
func (portabilityTransactionParticipantFake) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	return crossownertransaction.Input{SchemaID: "profile.import.v1", CanonicalBytes: []byte(`{"prepared":true}`)}, nil
}
func (portabilityTransactionParticipantFake) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{}, nil
}
func (portabilityTransactionParticipantFake) Validate(context.Context, crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	return crossownertransaction.Valid(), nil
}
func (portabilityTransactionParticipantFake) Write(context.Context, crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	return crossownertransaction.Written(nil), nil
}

func withPortabilityMode(policy ExtensionPolicy, mode string) ExtensionPolicy {
	policy.Mode = mode
	switch mode {
	case PortabilityModeNoAuthoritativeState:
		policy.ParticipantID = ""
		policy.BlockingFamilyIDs = nil
	case PortabilityModeBlockedWhenPresent:
		policy.ParticipantID = ""
		policy.BlockingFamilyIDs = append([]string(nil), policy.AuthoritativeFamilyIDs...)
	}
	return policy
}

func withClaim(policy ExtensionPolicy, claim string) ExtensionPolicy {
	policy.ClaimState = claim
	return policy
}

func replaceProfile(encoded []byte, old, replacement string) []byte {
	return []byte(strings.ReplaceAll(string(encoded), `"`+old+`"`, `"`+replacement+`"`))
}

func replaceMajor(encoded []byte, major int) []byte {
	return []byte(strings.Replace(string(encoded), `"contract_major":1`, `"contract_major":`+fmt.Sprint(major), 1))
}

func replaceDigest(encoded []byte, digest string) []byte {
	start := strings.Index(string(encoded), `"payload_sha256":"`) + len(`"payload_sha256":"`)
	result := append([]byte(nil), encoded...)
	copy(result[start:start+64], digest)
	return result
}
