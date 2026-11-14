package incidentbundles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
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
)

func TestIncidentBundlePortabilityStateAndClaimMatrix_Integration(t *testing.T) {
	t.Run("operation limits", checkPortabilityExportLimits)
	incidentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	base := ExtensionPolicy{
		ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v1",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
		Export:                 PortabilityOperationPolicy{OutputSchemaID: "profile.state.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
		Import:                 PortabilityOperationPolicy{OutputSchemaID: "profile.import.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
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
			wantFiles := map[string][]byte{}
			if test.wantPayload {
				data := []byte("{\"value\":\"portable\"}\n")
				dataPath := "ext/extensions/profile/members/state.ndjson"
				catalog, err := json.Marshal(map[string]any{
					"schema_id":  "cartulary.incident_bundle_extension_members.v2",
					"profile_id": "profile", "contract_major": 1, "state_version": 1,
					"content_schema_id": "profile.state.v1",
					"members":           []map[string]any{{"path": dataPath, "byte_size": len(data), "sha256": fmt.Sprintf("%x", sha256.Sum256(data))}},
				})
				if err != nil {
					t.Fatal(err)
				}
				wantFiles[dataPath] = data
				wantFiles["ext/extensions/profile/catalog.json"] = append(catalog, '\n')
			}
			wantMembers := make([]incidentportability.Member, 0, len(wantFiles))
			for path, want := range wantFiles {
				wantMembers = append(wantMembers, incidentportability.Member{Path: path, ByteSize: int64(len(want)), SHA256: fmt.Sprintf("%x", sha256.Sum256(want))})
				got, err := incidentportability.ReadControl(t.Context(), members, path, PortabilityParticipantByteLimit)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatalf("member %s bytes=%q want=%q err=%v", path, got, want, err)
				}
			}
			slices.SortFunc(wantMembers, func(a, b incidentportability.Member) int { return strings.Compare(a.Path, b.Path) })
			if got := members.Members(); !slices.Equal(got, wantMembers) {
				t.Fatalf("members=%+v want=%+v", got, wantMembers)
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
	t.Run("operation limits", checkPortabilityImportLimits)
	ctx := t.Context()
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	input := PortabilityImportContext{OperationID: "operation", IncidentID: incidentID, ActorUserID: uuid.MustParse("33333333-3333-4333-8333-333333333333")}
	policy := ExtensionPolicy{ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v2",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
		Export:                 PortabilityOperationPolicy{OutputSchemaID: "profile.state.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
		Import:                 PortabilityOperationPolicy{OutputSchemaID: "profile.import.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
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
	if _, exists := actors.ids["44444444-4444-4444-8444-444444444444"]; !exists {
		t.Fatalf("collected actors=%v", actors.ids)
	}
	members.Seal()
	prepared, err := orchestrator.PrepareImport(ctx, input, members)
	if err != nil {
		t.Fatal(err)
	}
	if participant.importCalls != 1 || len(prepared.Participants) != 1 || len(prepared.Publications) != 1 {
		t.Fatal("participant or transfer missing")
	}
	if got := prepared.Participants[0].ID(); got != "profile.import_v1" {
		t.Fatalf("prepared participant=%q", got)
	}
	transfer := prepared.Publications[0].Transfer
	if transfer.OperationID() != "operation" || transfer.ProfileID() != "profile" || !slices.Equal(transfer.References(), []string{"operation:profile"}) {
		t.Fatalf("transfer operation=%q profile=%q refs=%v", transfer.OperationID(), transfer.ProfileID(), transfer.References())
	}
	if !slices.Equal(participant.stagedRefs, []string{"cartulary:staged_output:operation:profile"}) || !slices.Equal(allocator.allocated, []string{"operation:profile"}) {
		t.Fatalf("logical refs=%v allocations=%v", participant.stagedRefs, allocator.allocated)
	}
	wantPublications := []stagedobjects.Publication{{
		StagingID: "operation:profile", ResourceKind: "test_resource", ResourceID: "22222222-2222-2222-2222-222222222222",
		ByteSize: 6, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("staged"))),
	}}
	if got := prepared.Publications[0].Publications; !slices.Equal(got, wantPublications) {
		t.Fatalf("publications=%+v want=%+v", got, wantPublications)
	}
	bound, err := prepared.Participants[0].BuildInput(ctx, crossownertransaction.OperationContext{OperationID: input.OperationID})
	if err != nil || string(bound.CanonicalBytes) != "{\"prepared\":true}\n" {
		t.Fatalf("bound input: %v", err)
	}
	if bound.SchemaID != "profile.import.v1" {
		t.Fatalf("prepared input schema=%q", bound.SchemaID)
	}
	if _, err := prepared.Participants[0].BuildInput(ctx, crossownertransaction.OperationContext{OperationID: "another"}); err == nil {
		t.Fatal("prepared descriptor crossed operations")
	}
	if err := prepared.Abandon(ctx); err != nil {
		t.Fatalf("abandon: %v", err)
	}
	if !slices.Equal(allocator.abandoned, []string{"operation:profile"}) {
		t.Fatalf("abandoned=%v", allocator.abandoned)
	}
	catalogPath := extensionCatalogPath("profile")
	raw, err := incidentportability.ReadControl(ctx, members, catalogPath, PortabilityParticipantByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	dataPath := extensionMemberPrefix("profile") + "state.ndjson"
	for _, tc := range []struct {
		name          string
		policy        ExtensionPolicy
		mutate        func(map[string][]byte)
		wantError     error
		wantInvariant string
	}{
		{"unknown profile", policy, func(files map[string][]byte) {
			delete(files, catalogPath)
			unknown := bytes.Replace(raw, []byte(`"profile_id":"profile"`), []byte(`"profile_id":"unknown"`), 1)
			files[extensionCatalogPath("unknown")] = bytes.ReplaceAll(unknown, []byte(extensionMemberPrefix("profile")), []byte(extensionMemberPrefix("unknown")))
			files[extensionMemberPrefix("unknown")+"state.ndjson"] = files[dataPath]
			delete(files, dataPath)
		}, ErrPortabilityUnavailable, "extension_payload.participant_admitted"},
		{"recognized unclaimable", withClaim(policy, ClaimStateRecognizedUnclaimable), nil, ErrPortabilityUnavailable, "extension_payload.participant_admitted"},
		{"unclaimed", withClaim(policy, ClaimStateUnclaimed), nil, ErrPortabilityUnavailable, "extension_payload.participant_admitted"},
		{"nonparticipant", withPortabilityMode(policy, PortabilityModeNoAuthoritativeState), nil, ErrPortabilityUnavailable, "extension_payload.participant_admitted"},
		{"incompatible major", policy, func(files map[string][]byte) {
			files[catalogPath] = bytes.Replace(raw, []byte("\"contract_major\":1"), []byte("\"contract_major\":2"), 1)
		}, ErrPortabilityResult, "extension_payload.contract_compatible"},
		{"wrong schema", policy, func(files map[string][]byte) {
			files[catalogPath] = bytes.Replace(raw, []byte("profile.state.v1"), []byte("profile.other.v1"), 1)
		}, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
		{"digest mismatch", policy, func(files map[string][]byte) { files[dataPath] = []byte("changed") }, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
		{"undeclared member", policy, func(files map[string][]byte) { files[extensionMemberPrefix("profile")+"extra"] = []byte("undeclared") }, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
		{"missing member", policy, func(files map[string][]byte) { delete(files, dataPath) }, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
		{"inline predecessor", policy, func(files map[string][]byte) {
			delete(files, catalogPath)
			files["ext/extensions/profile/payload.json"] = []byte(`{"payload":{}}`)
		}, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
		{"noncanonical catalog", policy, func(files map[string][]byte) { files[catalogPath] = append([]byte(" "), raw...) }, ErrPortabilityPayload, "extension_payload.schema_digest_valid"},
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
			_, err = current.PrepareImport(ctx, input, fixtureMembers(t, files))
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("admission error=%v, want %v", err, tc.wantError)
			}
			reason, details := incidentBundleFailureDetails("incident_bundle_import_rejected", err)
			if reason != "source_family_invalid" || !reflect.DeepEqual(details, map[string]any{
				"reason_code": "source_family_invalid", "source_family_id": "extension_payload",
				"invariant_id": tc.wantInvariant,
			}) {
				t.Fatalf("admission reason=%q details=%#v", reason, details)
			}
			if fake.importCalls != 0 {
				t.Fatal("participant ran before admission")
			}
		})
	}
	for _, tc := range []struct {
		name        string
		mutate      func(map[string]any, map[string]any)
		wantInvalid bool
	}{
		{name: "zero-byte member"},
		{name: "missing member size", wantInvalid: true, mutate: func(_ map[string]any, member map[string]any) { delete(member, "byte_size") }},
		{name: "null member size", wantInvalid: true, mutate: func(_ map[string]any, member map[string]any) { member["byte_size"] = nil }},
		{name: "member field case alias", wantInvalid: true, mutate: func(_ map[string]any, member map[string]any) { delete(member, "byte_size"); member["BYTE_SIZE"] = 0 }},
		{name: "catalog field case alias", wantInvalid: true, mutate: func(catalog map[string]any, _ map[string]any) {
			catalog["SCHEMA_ID"] = catalog["schema_id"]
			delete(catalog, "schema_id")
		}},
		{name: "unknown member field", wantInvalid: true, mutate: func(_ map[string]any, member map[string]any) { member["extra"] = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			member := map[string]any{"path": dataPath, "byte_size": 0, "sha256": fmt.Sprintf("%x", sha256.Sum256(nil))}
			catalog := map[string]any{
				"schema_id": "cartulary.incident_bundle_extension_members.v2", "profile_id": "profile",
				"contract_major": 1, "state_version": 1, "content_schema_id": "profile.state.v1",
				"members": []map[string]any{member},
			}
			if tc.mutate != nil {
				tc.mutate(catalog, member)
			}
			raw, err := json.Marshal(catalog)
			if err != nil {
				t.Fatal(err)
			}
			fake := &portabilityParticipantFake{id: policy.ParticipantID}
			current, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{}, Participants: []ExtensionParticipant{fake}, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{}})
			if err != nil {
				t.Fatal(err)
			}
			prepared, err := current.PrepareImport(t.Context(), input, fixtureMembers(t, map[string][]byte{catalogPath: append(raw, '\n'), dataPath: {}}))
			if tc.wantInvalid {
				if !errors.Is(err, ErrPortabilityPayload) || fake.importCalls != 0 {
					t.Fatalf("invalid catalog: error=%v participant calls=%d", err, fake.importCalls)
				}
				return
			}
			if err != nil || fake.importCalls != 1 || fake.readBytes != 0 {
				t.Fatalf("zero-byte member: error=%v calls=%d bytes=%d", err, fake.importCalls, fake.readBytes)
			}
			if err := prepared.Abandon(t.Context()); err != nil {
				t.Fatal(err)
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
	const wantInput = `{"incident_id":"12345678-1234-4234-8234-123456789abc","job_id":"00000000-0000-0000-0000-000000000000","manifest_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","schema_id":"cartulary.incident_portability_import_transaction_input.v2"}` + "\n"
	if produced.SchemaID != "cartulary.incident_portability_import_transaction_input.v2" || string(produced.CanonicalBytes) != wantInput {
		t.Fatalf("import input schema=%q bytes=%q, want %q", produced.SchemaID, produced.CanonicalBytes, wantInput)
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
	const contract = `{"owned_state_family_ids":["core.incident","incident_portability.job"],"owner_profile_id":"incident_portability","participant_id":"incident_portability.import_v2","participant_input_schema_id":"cartulary.incident_portability_import_transaction_input.v2","prepare_algorithm_id":"incident_portability.import_prepare_v2","schema_id":"cartulary.extension_transaction_participant_contract.v3","serialization_key_kinds":["incident_portability.incident"],"validation_algorithm_id":"incident_portability.import_validate_v2","write_algorithm_id":"incident_portability.import_write_v2"}` + "\n"
	wantDescriptor := crossownertransaction.Descriptor{
		ParticipantID:         "incident_portability.import_v2",
		OwnerProfileID:        "incident_portability",
		ContractSHA256:        fmt.Sprintf("%x", sha256.Sum256([]byte(contract))),
		InputSchemaID:         "cartulary.incident_portability_import_transaction_input.v2",
		PrepareAlgorithmID:    "incident_portability.import_prepare_v2",
		ValidationAlgorithmID: "incident_portability.import_validate_v2",
		WriteAlgorithmID:      "incident_portability.import_write_v2",
		SerializationKeyKinds: []string{"incident_portability.incident"},
		OwnedStateFamilyIDs:   []string{"core.incident", "incident_portability.job"},
	}
	if !reflect.DeepEqual(descriptor, wantDescriptor) {
		t.Fatalf("import descriptor = %#v, want %#v", descriptor, wantDescriptor)
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
func (portabilityReadFactoryFake) ForSnapshot(query ExportSnapshotQuery, incident uuid.UUID, profile string, _ []string) (ExportReadCapability, error) {
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
