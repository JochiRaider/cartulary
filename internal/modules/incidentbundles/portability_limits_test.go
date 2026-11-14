package incidentbundles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
)

func portabilityLimitJSON(t *testing.T, value any) []byte {
	t.Helper()
	// All fixture keys and values are ASCII. Standard JSON independently orders
	// these maps; the framing byte is part of the expected metadata budget.
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return append(raw, '\n')
}

func portabilityLimitPolicy() ExtensionPolicy {
	return ExtensionPolicy{
		ProfileID: "profile", ClaimState: "claimed", ContractMajor: 1, Mode: "participant",
		ParticipantID: "profile.portability", ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
		Export:                 PortabilityOperationPolicy{OutputSchemaID: "profile.state.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: 67108864, MaximumOutputBytes: 67108864, MaximumItems: 10000},
		Import:                 PortabilityOperationPolicy{OutputSchemaID: "profile.import.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: 67108864, MaximumOutputBytes: 67108864, MaximumItems: 10000},
		AuthoritativeFamilyIDs: []string{"profile.state"},
	}
}

type portabilityBudgetParticipant struct {
	*portabilityParticipantFake
	exportMembers, preparationMembers, publicationMembers int
	streamPublication                                     bool
}

type portabilityBudgetReadFactory struct{ selections [][]string }

func (f *portabilityBudgetReadFactory) ForSnapshot(query ExportSnapshotQuery, incident uuid.UUID, profile string, families []string) (ExportReadCapability, error) {
	f.selections = append(f.selections, slices.Clone(families))
	return (portabilityReadFactoryFake{}).ForSnapshot(query, incident, profile, families)
}

func (p *portabilityBudgetParticipant) Export(ctx context.Context, input ExportInvocation) (ExportResult, error) {
	p.exportCalls++
	for index := range p.exportMembers {
		if _, err := input.Members.WriteMember(ctx, fmt.Sprintf("ext/extensions/profile/members/%d.ndjson", index), bytes.NewReader(nil)); err != nil {
			return ExportResult{}, err
		}
	}
	return ExportResult{SchemaID: "cartulary.extension_portability_export_result.v2", Kind: "members", ContentSchemaID: "profile.state.v1", ContractMajor: 1, StateVersion: 1}, nil
}

func (p *portabilityBudgetParticipant) PrepareImport(ctx context.Context, input ImportInvocation, scope *PortabilityStagedOutputScope) (ImportPreparation, error) {
	p.importCalls++
	for index := range p.preparationMembers {
		if _, err := input.PreparationMembers.WriteMember(ctx, fmt.Sprintf("prepared/profile/%d.ndjson", index), bytes.NewReader(nil)); err != nil {
			return ImportPreparation{}, err
		}
	}
	for range p.publicationMembers {
		resource := stagedobjects.ResourceBinding{Kind: "test_resource", ID: input.IncidentID.String()}
		if p.streamPublication {
			if _, _, err := scope.AllocateStream(ctx, input.OperationID, resource, bytes.NewReader(nil), 1); err != nil {
				return ImportPreparation{}, err
			}
		} else if _, err := scope.Allocate(ctx, input.OperationID, resource, nil); err != nil {
			return ImportPreparation{}, err
		}
	}
	raw := []byte("{\"prepared\":true}\n")
	return ImportPreparation{SchemaID: "cartulary.extension_portability_import_preparation_result.v2", Status: "prepared", ParticipantInput: raw,
		ParticipantInputSHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), StagedOutputRefs: scope.Refs(), TransactionParticipant: portabilityTransactionParticipantFake{}}, nil
}

func checkPortabilityExportLimits(t *testing.T) {
	incident := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	contextBytes := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.extension_portability_participant_context.v2", "operation_kind": "export", "profile_id": "profile", "contract_major": 1, "claim_state": "claimed", "incident_id": incident.String(), "content_schema_id": "profile.state.v1"})
	members := []map[string]any{}
	for index := range 2 {
		members = append(members, map[string]any{"path": fmt.Sprintf("ext/extensions/profile/members/%d.ndjson", index), "byte_size": 0, "sha256": fmt.Sprintf("%x", sha256.Sum256(nil))})
	}
	catalog := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.incident_bundle_extension_members.v2", "profile_id": "profile", "contract_major": 1, "state_version": 1, "content_schema_id": "profile.state.v1", "members": members})
	result := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.extension_portability_export_result.v2", "kind": "members", "content_schema_id": "profile.state.v1", "contract_major": 1, "state_version": 1})
	for _, tc := range []struct {
		name                                           string
		inputBytes, outputBytes                        int
		families, maximumItems, wantCalls, wantMembers int
		limited                                        bool
	}{
		{"exact metadata and items", len(contextBytes), len(result) + len(catalog), 2, 2, 1, 3, false},
		{"context one byte over", len(contextBytes) - 1, len(result) + len(catalog), 2, 2, 0, 0, true},
		{"result plus catalog one byte over", len(contextBytes), len(result) + len(catalog) - 1, 2, 2, 1, 2, true},
		{"input item overflow", len(contextBytes), len(result) + len(catalog), 2, 1, 0, 0, true},
		{"output item overflow", len(contextBytes), len(result) + len(catalog), 1, 1, 1, 1, true},
		{"zero selected items", len(contextBytes), len(result) + len(catalog), 1, 0, 0, 0, true},
		{"zero output bytes", len(contextBytes), 0, 1, 2, 1, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := portabilityLimitPolicy()
			policy.Export.MaximumInputBytes, policy.Export.MaximumOutputBytes, policy.Export.MaximumItems = int64(tc.inputBytes), int64(tc.outputBytes), tc.maximumItems
			if tc.families == 2 {
				policy.Export.StateFamilyIDs = []string{"profile.first", "profile.second"}
				policy.AuthoritativeFamilyIDs = slices.Clone(policy.Export.StateFamilyIDs)
			}
			participant := &portabilityBudgetParticipant{portabilityParticipantFake: &portabilityParticipantFake{id: policy.ParticipantID}, exportMembers: 2}
			factory := &portabilityBudgetReadFactory{}
			orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true}, Participants: []ExtensionParticipant{participant}, SnapshotReads: factory})
			if err != nil {
				t.Fatal(err)
			}
			output := newArchiveTestMembers(t, Limits{})
			err = orchestrator.Export(t.Context(), portabilityQueryFake{}, incident, output, newActorReferenceSet())
			if errors.Is(err, ErrPortabilityLimit) != tc.limited || (!tc.limited && err != nil) || participant.exportCalls != tc.wantCalls || len(output.Members()) != tc.wantMembers {
				t.Fatalf("error=%v calls=%d members=%d; want limit=%v calls=%d members=%d", err, participant.exportCalls, len(output.Members()), tc.limited, tc.wantCalls, tc.wantMembers)
			}
			wantFamilies := []string{"profile.state"}
			if tc.families == 2 {
				wantFamilies = []string{"profile.first", "profile.second"}
			}
			if len(factory.selections) != tc.wantCalls || tc.wantCalls == 1 && !slices.Equal(factory.selections[0], wantFamilies) {
				t.Fatalf("scoped input selections=%v; want calls=%d families=%v", factory.selections, tc.wantCalls, wantFamilies)
			}
		})
	}
}

func checkPortabilityImportLimits(t *testing.T) {
	input := PortabilityImportContext{OperationID: "22222222-2222-4222-8222-222222222222", IncidentID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), ActorUserID: uuid.MustParse("33333333-3333-4333-8333-333333333333")}
	path := "ext/extensions/profile/members/state.ndjson"
	catalog := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.incident_bundle_extension_members.v2", "profile_id": "profile", "contract_major": 1, "state_version": 1, "content_schema_id": "profile.state.v1", "members": []map[string]any{{"path": path, "byte_size": 0, "sha256": fmt.Sprintf("%x", sha256.Sum256(nil))}}})
	contextBytes := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.extension_portability_participant_context.v2", "operation_kind": "import", "profile_id": "profile", "contract_major": 1, "claim_state": "claimed", "incident_id": input.IncidentID.String(), "content_schema_id": "profile.state.v1", "operation_id": input.OperationID, "actor_user_id": input.ActorUserID.String(), "state_version": 1, "catalog_sha256": fmt.Sprintf("%x", sha256.Sum256(catalog))})
	raw := []byte("{\"prepared\":true}\n")
	result := portabilityLimitJSON(t, map[string]any{"schema_id": "cartulary.extension_portability_import_preparation_result.v2", "status": "prepared", "participant_input": json.RawMessage(raw), "participant_input_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "staged_output_refs": []string{"cartulary:staged_output:2"}})
	publications := portabilityLimitJSON(t, []map[string]any{{"StagingID": "2", "ResourceKind": "test_resource", "ResourceID": input.IncidentID.String(), "ByteSize": 0, "SHA256": fmt.Sprintf("%x", sha256.Sum256(nil))}})
	for _, tc := range []struct {
		name                                                              string
		inputBytes, outputBytes, maximumItems, wantCalls, wantAllocations int
		preparationMembers, publicationMembers                            int
		streamPublication, limited                                        bool
	}{
		{"exact metadata and combined items", len(contextBytes) + len(catalog), len(result) + len(publications), 2, 1, 2, 1, 1, false, false},
		{"context plus catalog one byte over", len(contextBytes) + len(catalog) - 1, len(result) + len(publications), 2, 0, 0, 1, 1, false, true},
		{"result plus publication one byte over", len(contextBytes) + len(catalog), len(result) + len(publications) - 1, 2, 1, 2, 1, 1, false, true},
		{"zero input items", len(contextBytes) + len(catalog), len(result) + len(publications), 0, 0, 0, 1, 1, false, true},
		{"combined output item overflow", len(contextBytes) + len(catalog), len(result) + len(publications), 1, 1, 1, 1, 1, false, true},
		{"zero output bytes", len(contextBytes) + len(catalog), 0, 2, 1, 2, 1, 1, false, true},
		{"preparation output item overflow", len(contextBytes) + len(catalog), len(result) + len(publications), 1, 1, 1, 2, 0, false, true},
		{"stream publication output item overflow", len(contextBytes) + len(catalog), len(result) + len(publications), 1, 1, 1, 1, 1, true, true},
		{"stream publication exact limit", len(contextBytes) + len(catalog), len(result) + len(publications), 2, 1, 2, 1, 1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := portabilityLimitPolicy()
			policy.Import.MaximumInputBytes, policy.Import.MaximumOutputBytes, policy.Import.MaximumItems = int64(tc.inputBytes), int64(tc.outputBytes), tc.maximumItems
			participant := &portabilityBudgetParticipant{portabilityParticipantFake: &portabilityParticipantFake{id: policy.ParticipantID}, preparationMembers: tc.preparationMembers, publicationMembers: tc.publicationMembers, streamPublication: tc.streamPublication}
			allocator := newWorkerArchiveStaging(t)
			orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{}, Participants: []ExtensionParticipant{participant}, SnapshotReads: portabilityReadFactoryFake{}, Allocator: allocator})
			if err != nil {
				t.Fatal(err)
			}
			prepared, err := orchestrator.PrepareImport(t.Context(), input, fixtureMembers(t, map[string][]byte{"ext/extensions/profile/catalog.json": catalog, path: {}}))
			if errors.Is(err, ErrPortabilityLimit) != tc.limited || (!tc.limited && err != nil) || participant.importCalls != tc.wantCalls || allocator.disk.next != tc.wantAllocations {
				t.Errorf("error=%v calls=%d allocations=%d; want limit=%v calls=%d allocations=%d", err, participant.importCalls, allocator.disk.next, tc.limited, tc.wantCalls, tc.wantAllocations)
			}
			if err == nil {
				if abandonErr := prepared.Abandon(t.Context()); abandonErr != nil {
					t.Fatal(abandonErr)
				}
			}
			if len(allocator.disk.refs) != 0 {
				t.Fatalf("operation retained %d members after rejection/abandon", len(allocator.disk.refs))
			}
		})
	}
}
