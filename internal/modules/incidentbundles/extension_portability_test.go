package incidentbundles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestIncidentBundlePortabilityStateAndClaimMatrix_Integration(t *testing.T) {
	incidentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	base := ExtensionPolicy{
		ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v1",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
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
			var participants []ExtensionParticipant
			if test.policy.Mode == PortabilityModeParticipant && test.policy.ClaimState == ClaimStateClaimed {
				participants = []ExtensionParticipant{participant}
			}
			orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{test.policy}, Presence: portabilityPresenceFake{present: test.present}, Participants: participants, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{}})
			if err != nil {
				t.Fatal(err)
			}
			members := newArchiveTestMembers(t, Limits{})
			err = orchestrator.Export(t.Context(), portabilityQueryFake{}, incidentID, members, newActorReferenceSet())
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error=%v want=%v", err, test.wantError)
			}
			if got := len(members.Members()) == 2; got != test.wantPayload {
				t.Fatalf("members=%v", members.Members())
			}
			if participant.exportCalls != test.wantCalls {
				t.Fatalf("participant calls=%d want=%d", participant.exportCalls, test.wantCalls)
			}
		})
	}
	t.Run("blocking profiles use canonical profile order", func(t *testing.T) {
		later := withPortabilityMode(base, PortabilityModeBlockedWhenPresent)
		later.ProfileID = "z_profile"
		earlier := withPortabilityMode(base, PortabilityModeBlockedWhenPresent)
		earlier.ProfileID = "a_profile"
		orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{later, earlier}, Presence: portabilityPresenceFake{present: true}, Allocator: &portabilityAllocatorFake{}})
		if err != nil {
			t.Fatal(err)
		}
		err = orchestrator.Export(t.Context(), portabilityQueryFake{}, incidentID, newArchiveTestMembers(t, Limits{}), newActorReferenceSet())
		var failure *PortabilityFailure
		if !errors.As(err, &failure) || !errors.Is(err, ErrPortabilityBlocked) || failure.ProfileID != "a_profile" {
			t.Fatalf("blocking order: %v", err)
		}
		if err := orchestrator.ValidatePublication(t.Context(), portabilityQueryFake{}, incidentID); !errors.As(err, &failure) || failure.ProfileID != "a_profile" {
			t.Fatalf("publication guard: %v", err)
		}
	})
}

func TestIncidentBundlePortabilityImportAdmissionAndCleanup_Integration(t *testing.T) {
	ctx := t.Context()
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	input := PortabilityImportContext{OperationID: "operation", IncidentID: incidentID, ActorUserID: uuid.MustParse("33333333-3333-4333-8333-333333333333")}
	policy := ExtensionPolicy{ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v2",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
		MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit,
		AuthoritativeFamilyIDs: []string{"profile.state"}}
	allocator := &portabilityAllocatorFake{}
	participant := &portabilityParticipantFake{id: policy.ParticipantID, allocate: true}
	orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true}, Participants: []ExtensionParticipant{participant}, Allocator: allocator, SnapshotReads: portabilityReadFactoryFake{}})
	if err != nil {
		t.Fatal(err)
	}
	members := newArchiveTestMembers(t, Limits{})
	actors := newActorReferenceSet()
	if err := orchestrator.Export(ctx, portabilityQueryFake{}, incidentID, members, actors); err != nil {
		t.Fatal(err)
	}
	if len(actors.ids) != 1 {
		t.Fatal("extension-only actor was not collected before roster closure")
	}
	members.Seal()
	prepared, err := orchestrator.PrepareImport(ctx, input, members)
	if err != nil {
		t.Fatal(err)
	}
	if participant.importCalls != 1 || len(prepared.Participants) != 1 || len(prepared.Publications) != 1 {
		t.Fatal("participant or transfer missing")
	}
	if refs := prepared.Publications[0].Transfer.References(); len(refs) != 1 || !strings.HasPrefix(participant.stagedRefs[0], "cartulary:staged_output:") {
		t.Fatal("staging identity escaped logical reference")
	}
	bound, err := prepared.Participants[0].BuildInput(ctx, crossownertransaction.OperationContext{OperationID: input.OperationID})
	if err != nil || string(bound.CanonicalBytes) != "{\"prepared\":true}\n" {
		t.Fatalf("bound input: %v", err)
	}
	if _, err := prepared.Participants[0].BuildInput(ctx, crossownertransaction.OperationContext{OperationID: "another"}); err == nil {
		t.Fatal("prepared descriptor crossed operations")
	}
	if err := prepared.Abandon(ctx); err != nil || len(allocator.abandoned) != 1 {
		t.Fatalf("abandon: %v", err)
	}
	catalogPath := extensionCatalogPath("profile")
	raw, err := incidentportability.ReadControl(ctx, members, catalogPath, PortabilityParticipantByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	dataPath := extensionMemberPrefix("profile") + "state.ndjson"
	for _, tc := range []struct {
		name   string
		policy ExtensionPolicy
		mutate func(map[string][]byte)
	}{
		{"unknown profile", policy, func(files map[string][]byte) {
			delete(files, catalogPath)
			files[extensionCatalogPath("unknown")] = bytes.ReplaceAll(raw, []byte("profile"), []byte("unknown"))
		}},
		{"recognized unclaimable", withClaim(policy, ClaimStateRecognizedUnclaimable), nil},
		{"unclaimed", withClaim(policy, ClaimStateUnclaimed), nil},
		{"nonparticipant", withPortabilityMode(policy, PortabilityModeNoAuthoritativeState), nil},
		{"incompatible major", policy, func(files map[string][]byte) {
			files[catalogPath] = bytes.Replace(raw, []byte("\"contract_major\":1"), []byte("\"contract_major\":2"), 1)
		}},
		{"digest mismatch", policy, func(files map[string][]byte) { files[dataPath] = []byte("changed") }},
		{"undeclared member", policy, func(files map[string][]byte) { files[extensionMemberPrefix("profile")+"extra"] = []byte("undeclared") }},
		{"missing member", policy, func(files map[string][]byte) { delete(files, dataPath) }},
		{"inline predecessor", policy, func(files map[string][]byte) {
			delete(files, catalogPath)
			files["ext/extensions/profile/payload.json"] = []byte(`{"payload":{}}`)
		}},
		{"noncanonical catalog", policy, func(files map[string][]byte) { files[catalogPath] = append([]byte(" "), raw...) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string][]byte{catalogPath: raw, dataPath: []byte("{\"value\":\"portable\"}\n")}
			if tc.mutate != nil {
				tc.mutate(files)
			}
			var participants []ExtensionParticipant
			fake := &portabilityParticipantFake{id: policy.ParticipantID}
			if tc.policy.Mode == PortabilityModeParticipant && tc.policy.ClaimState == ClaimStateClaimed {
				participants = []ExtensionParticipant{fake}
			}
			current, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{tc.policy}, Presence: portabilityPresenceFake{}, Participants: participants, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := current.PrepareImport(ctx, input, fixtureMembers(t, files)); err == nil {
				t.Fatal("invalid member catalog admitted")
			}
			if fake.importCalls != 0 {
				t.Fatal("participant ran before admission")
			}
		})
	}
	// A bounded bulk member exceeding the old inline ceiling remains streamed.
	participant.bulkBytes = 65 << 20
	bulk := newArchiveTestMembers(t, Limits{})
	if err := orchestrator.Export(ctx, portabilityQueryFake{}, incidentID, bulk, newActorReferenceSet()); err != nil {
		t.Fatal(err)
	}
	bulk.Seal()
	prepared, err = orchestrator.PrepareImport(ctx, input, bulk)
	if err != nil {
		t.Fatal(err)
	}
	if participant.readBytes != participant.bulkBytes {
		t.Fatalf("bulk bytes=%d", participant.readBytes)
	}
	if err := prepared.Abandon(ctx); err != nil {
		t.Fatal(err)
	}
	// The control ceiling still applies independently of admitted bulk bytes.
	tooLarge := newArchiveTestMembers(t, Limits{})
	if _, err := tooLarge.WriteMember(ctx, catalogPath, io.LimitReader(zeroStream{}, PortabilityParticipantByteLimit+1)); err != nil {
		t.Fatal(err)
	}
	if _, err := orchestrator.PrepareImport(ctx, input, tooLarge); !errors.Is(err, ErrPortabilityLimit) {
		t.Fatalf("control overflow: %v", err)
	}
}

type zeroStream struct{}

func (zeroStream) Read(p []byte) (int, error) {
	if len(p) > 64<<10 {
		return 0, fmt.Errorf("bulk read exceeds bounded chunk")
	}
	clear(p)
	return len(p), nil
}

type importCapabilityShapeTx struct{ pgx.Tx }

func TestIncidentBundleImportDescriptorIsClosed_Unit(t *testing.T) {
	module := &Module{store: &store{}, jobs: jobAdmissionStub{}}
	read, write, err := module.TransactionCapabilities(ImportTransactionParticipantID, &importCapabilityShapeTx{})
	if err != nil || read.ParticipantScope() != ImportTransactionParticipantID || write.ParticipantScope() != ImportTransactionParticipantID {
		t.Fatalf("real owner capabilities: %v", err)
	}
	if _, ok := read.(importReadCapability); !ok {
		t.Fatal("reader lacks its logical check")
	}
	if _, ok := write.(importWriteCapability); !ok {
		t.Fatal("writer lacks its logical mutation")
	}
	if _, ok := read.(importWriteCapability); ok {
		t.Fatal("reader exposes bundle mutation")
	}
	if _, ok := write.(importReadCapability); ok {
		t.Fatal("writer exposes validation")
	}
	if reflect.TypeOf(read).NumMethod() != 2 || reflect.TypeOf(write).NumMethod() != 2 {
		t.Fatal("physical factory exposes methods outside the selected phase")
	}

	// Encoding is independent of the job lease; the schema fixture below supplies
	// a valid logical job identity. Live execution admission remains Jobs-owned.
	producer := &importTransactionParticipant{prepared: &preparedImport{IncidentID: uuid.MustParse("12345678-1234-4234-8234-123456789abc")}, manifestSHA: strings.Repeat("a", 64)}
	produced, produceErr := producer.BuildInput(context.Background(), crossownertransaction.OperationContext{})
	if produceErr != nil {
		t.Fatal(produceErr)
	}
	if _, err := canonicaljson.DecodeExtension(produced.CanonicalBytes); err != nil || produced.SchemaID != importTransactionInputSchema {
		t.Fatalf("noncanonical import producer: %v", err)
	}

	input := map[string]any{"schema_id": importTransactionInputSchema, "incident_id": "12345678-1234-4234-8234-123456789abc", "job_id": "22345678-1234-4234-8234-123456789abc", "manifest_sha256": strings.Repeat("a", 64)}
	if !ValidateImportTransactionInput(input) {
		t.Fatal("valid import transaction input rejected")
	}
	input["schema_id"] = "cartulary.incident_portability_import_transaction_input.v1"
	if ValidateImportTransactionInput(input) {
		t.Fatal("retired input admitted")
	}
	input["schema_id"] = importTransactionInputSchema
	for _, key := range []string{"schema_id", "incident_id", "job_id", "manifest_sha256", "unknown"} {
		copy := map[string]any{}
		for k, v := range input {
			copy[k] = v
		}
		copy[key] = "invalid"
		if ValidateImportTransactionInput(copy) {
			t.Fatalf("invalid %s input admitted", key)
		}
	}

	descriptor := ImportTransactionDescriptor()
	if descriptor.ParticipantID != ImportTransactionParticipantID ||
		len(descriptor.SerializationKeyKinds) != 1 ||
		len(descriptor.OwnedStateFamilyIDs) != 2 ||
		len(descriptor.ContractSHA256) != 64 {
		t.Fatalf("descriptor = %#v", descriptor)
	}
	contribution := RecoveryStateContribution()
	if contribution.OwnerID != "module.incidentbundles" || len(contribution.Tables) != 5 || len(contribution.ObjectFamilies) != 1 {
		t.Fatalf("root Recovery contribution identity changed: %#v", contribution)
	}
	wantTables := []string{
		"incident_bundle_exports",
		"incident_bundle_imported_actors",
		"incident_bundle_imported_attributions",
		"incident_bundle_job_payloads",
		"incident_bundle_manifest_files",
	}
	gotTables := make([]string, 0, len(contribution.Tables))
	for _, table := range contribution.Tables {
		gotTables = append(gotTables, table.TableName)
	}
	if !slices.Equal(gotTables, wantTables) {
		t.Fatalf("root Recovery table order = %#v; want %#v", gotTables, wantTables)
	}
	family := contribution.ObjectFamilies[0]
	if family.ObjectFamilyID != "incident_bundles.files" ||
		family.InventoryAlgorithmID == nil || *family.InventoryAlgorithmID != "incidentbundles.snapshot_file_inventory.v1" ||
		family.ValidationAlgorithmID == nil || *family.ValidationAlgorithmID != "incidentbundles.validate_file_inventory.v1" ||
		family.RestoreAlgorithmID == nil || *family.RestoreAlgorithmID != "incidentbundles.restore_file_inventory.v1" {
		t.Fatalf("root Recovery object family changed: %#v", family)
	}
	inventory := VNextRecoveryObjectInventory(nil)
	if inventory.OwnerID() != "module.incidentbundles" || inventory.ObjectFamilyID() != "incident_bundles.files" || inventory.InventoryAlgorithmID() != "incidentbundles.snapshot_file_inventory.v1" {
		t.Fatalf("root Recovery inventory identity changed: owner=%q family=%q algorithm=%q", inventory.OwnerID(), inventory.ObjectFamilyID(), inventory.InventoryAlgorithmID())
	}
}

type portabilityPresenceFake struct{ present bool }

func (p portabilityPresenceFake) AuthoritativeStatePresent(context.Context, StatePresenceQuery, uuid.UUID, string, []string) (bool, error) {
	return p.present, nil
}

type portabilityQueryFake struct{}

func (portabilityQueryFake) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unrestricted query must not reach participant")
}

func (portabilityQueryFake) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

type portabilityReadFactoryFake struct{}
type portabilityReadFake struct {
	profile  string
	incident uuid.UUID
}

func (r portabilityReadFake) ProfileScope() string     { return r.profile }
func (r portabilityReadFake) IncidentScope() uuid.UUID { return r.incident }
func (portabilityReadFactoryFake) ForSnapshot(query ExportSnapshotQuery, incident uuid.UUID, profile string) (ExportReadCapability, error) {
	if _, ok := query.(pgx.Tx); ok {
		return nil, errors.New("snapshot exposes transaction")
	}
	return portabilityReadFake{profile, incident}, nil
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
	bulkBytes   int64
	readBytes   int64
}

func (p *portabilityParticipantFake) ID() string { return p.id }

func (p *portabilityParticipantFake) ContractSHA256() string {
	return strings.Repeat("a", 64)
}

func (p *portabilityParticipantFake) SpecializationSchemaID() string {
	return "test.specialization.v2"
}

func (p *portabilityParticipantFake) Export(ctx context.Context, input ExportInvocation) (ExportResult, error) {
	p.exportCalls++
	if input.ReadAccess.ProfileScope() != input.ProfileID || input.ReadAccess.IncidentScope() != input.IncidentID {
		return ExportResult{}, ErrPortabilityResult
	}
	if _, ok := input.ReadAccess.(ExportSnapshotQuery); ok {
		return ExportResult{}, errors.New("participant received SQL")
	}
	if err := input.Actors.RecordActorReference(uuid.MustParse("44444444-4444-4444-8444-444444444444")); err != nil {
		return ExportResult{}, err
	}
	var source io.Reader = strings.NewReader("{\"value\":\"portable\"}\n")
	if p.bulkBytes > 0 {
		source = io.LimitReader(zeroStream{}, p.bulkBytes)
	}
	if _, err := input.Members.WriteMember(ctx, extensionMemberPrefix(input.ProfileID)+"state.ndjson", source); err != nil {
		return ExportResult{}, err
	}
	return ExportResult{SchemaID: ExtensionExportResultSchema, Kind: "members", ContentSchemaID: "profile.state.v1", ContractMajor: 1, StateVersion: 1}, nil
}

func (p *portabilityParticipantFake) PrepareImport(ctx context.Context, invocation ImportInvocation, scope *PortabilityStagedOutputScope) (ImportPreparation, error) {
	p.importCalls++
	if invocation.ContentSchemaID != "profile.state.v1" || !validMemberSHA(invocation.CatalogSHA256) {
		return ImportPreparation{}, ErrPortabilityPayload
	}
	p.readBytes = 0
	for _, member := range invocation.Members.Members() {
		reader, err := incidentportability.OpenVerifiedMember(ctx, invocation.Members, member)
		if err != nil {
			return ImportPreparation{}, err
		}
		n, err := io.CopyBuffer(io.Discard, reader, make([]byte, 32<<10))
		closeErr := reader.Close()
		if err != nil {
			return ImportPreparation{}, err
		}
		if closeErr != nil {
			return ImportPreparation{}, closeErr
		}
		p.readBytes += n
	}
	if p.allocate {
		if _, err := scope.Allocate(ctx, invocation.OperationID, stagedobjects.ResourceBinding{Kind: "test_resource", ID: invocation.IncidentID.String()}, []byte("staged")); err != nil {
			return ImportPreparation{}, err
		}
	}
	p.stagedRefs = scope.Refs()
	input := []byte("{\"prepared\":true}\n")
	return ImportPreparation{
		SchemaID: ExtensionImportResultSchema, Status: "prepared",
		ParticipantInput: input, ParticipantInputSHA256: digestBytes(input),
		StagedOutputRefs: scope.Refs(), TransactionParticipant: portabilityTransactionParticipantFake{},
	}, nil
}

type portabilityTransactionParticipantFake struct{}

func (portabilityTransactionParticipantFake) ID() string { return "profile.import_v1" }
func (portabilityTransactionParticipantFake) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	return crossownertransaction.Input{SchemaID: "profile.import.v1", CanonicalBytes: []byte("{\"prepared\":true}\n")}, nil
}
func (portabilityTransactionParticipantFake) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{}}, nil
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
	if claim == ClaimStateRecognizedUnclaimable {
		policy.ContractMajor = 0
	}
	return policy
}
