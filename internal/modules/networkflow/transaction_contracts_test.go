package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func assertTransactionInputContracts(t *testing.T) {
	t.Run("retained import result admission", testRetainedImportResultAdmission)
	m := &Module{limits: defaultEffectiveLimits()}
	assertImportParticipantExecution(t, m)
	p := &indicatorLinkParticipant{mutation: indicatorLinkMutation{
		IncidentID: uuid.New(), Actor: authn.UserRecord{ID: uuid.New()},
		Request:  indicatorLinkRequest{ClientTxnID: "transaction", Selector: indicatorLinkSelector{Kind: "row_field_value", TableID: "nft_" + strings.Repeat("a", 32), RowID: "nfr_" + strings.Repeat("b", 64), FieldKey: fieldEndpointAIP}, Target: indicatorLinkTarget{Mode: "existing_indicator", IndicatorID: uuid.New()}},
		Resolved: resolvedIndicatorLinkSelector{CandidateValue: "192.0.2.10", SourceRowRefs: []networkFlowRowRef{{NetworkFlowTableID: "nft_" + strings.Repeat("a", 32), NetworkFlowRowID: "nfr_" + strings.Repeat("b", 64)}}},
	}}
	input, err := p.BuildInput(context.Background(), crossownertransaction.OperationContext{})
	if err != nil {
		t.Fatal(err)
	}
	value, err := canonicaljson.DecodeExtension(input.CanonicalBytes)
	if err != nil || !m.ValidateTransactionInput(input.SchemaID, value) {
		t.Fatalf("real input producer failed admission: %v", err)
	}
	for _, mutation := range []string{"unknown", "schema", "null", "actor", "candidate", "nested"} {
		value, _ := canonicaljson.DecodeExtension(input.CanonicalBytes)
		object := value.(map[string]any)
		switch mutation {
		case "unknown":
			object["private_extra"] = true
		case "schema":
			object["schema_id"] = "unknown"
		case "null":
			object["request"] = nil
		case "actor":
			object["actor_user_id"] = "invalid"
		case "candidate":
			object["candidate_value"] = "not-an-address"
		case "nested":
			object["target"].(map[string]any)["extra"] = true
		}
		if m.ValidateTransactionInput(input.SchemaID, object) {
			t.Fatalf("malformed %s input admitted", mutation)
		}
	}
	if m.ValidateTransactionInput("unknown", value) || (*Module)(nil).ValidateTransactionInput(input.SchemaID, value) {
		t.Fatal("unresolved input contract admitted")
	}
	oldValue, _ := canonicaljson.DecodeExtension(input.CanonicalBytes)
	oldValue.(map[string]any)["schema_id"] = "cartulary.network_flow_activity.indicator_link_transaction_input.v1"
	if m.ValidateTransactionInput("cartulary.network_flow_activity.indicator_link_transaction_input.v1", oldValue) {
		t.Fatal("retired input admitted")
	}
	prepared, err := p.Prepare(context.Background(), crossownertransaction.Invocation{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(prepared.SerializationKeys); i++ {
		a, b := prepared.SerializationKeys[i-1], prepared.SerializationKeys[i]
		if a.KeyKind > b.KeyKind || a.KeyKind == b.KeyKind && a.Key >= b.Key {
			t.Fatal("owner prepare result is not ordered")
		}
	}
	for _, id := range []string{ImportApplyParticipantID, IndicatorLinkParticipantID} {
		d, ok := TransactionContract(id)
		if !ok || d.ContractSHA256 != crossownertransaction.DescriptorDigest(d) {
			t.Fatal("unbound owner contract")
		}
	}
	if _, ok := TransactionContract("unknown"); ok {
		t.Fatal("unknown participant resolved")
	}
	mapping := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
	sourceDigest := strings.Repeat("a", 64)
	fingerprint, err := mappingFingerprint(mapping, sourceDigest)
	if err != nil {
		t.Fatal(err)
	}
	var mappingValue any
	if err := json.Unmarshal(marshalApprovedMapping(mapping), &mappingValue); err != nil {
		t.Fatal(err)
	}
	importValue := map[string]any{"schema_id": importApplyTransactionInputSchema, "incident_id": uuid.NewString(), "actor_user_id": uuid.NewString(), "import_session_id": uuid.NewString(), "import_unit_id": uuid.NewString(), "source_stream_ref": "opaque-owner-reference", "source_content_sha256": sourceDigest, "mapping_fingerprint": fingerprint, "owner_mapping_schema_id": importMappingSchemaID, "owner_mapping": mappingValue, "client_txn_id": "import-transaction"}
	if !m.ValidateTransactionInput(importApplyTransactionInputSchema, importValue) {
		t.Fatal("approved import input rejected")
	}
	importValue["mapping_fingerprint"] = strings.Repeat("b", 64)
	if m.ValidateTransactionInput(importApplyTransactionInputSchema, importValue) {
		t.Fatal("unbound approved mapping accepted")
	}

}

type importParticipantPortFixture struct {
	calls         []string
	validationErr error
}

func (*importParticipantPortFixture) ParticipantScope() string { return ImportApplyParticipantID }
func (p *importParticipantPortFixture) ValidateImportApply(context.Context, imports.ExtensionImportApplyRequest) error {
	p.calls = append(p.calls, "validate")
	return p.validationErr
}
func (p *importParticipantPortFixture) WriteImportApply(context.Context, imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error) {
	p.calls = append(p.calls, "write")
	return imports.ExtensionImportApplyResult{OwnerResponse: map[string]any{"schema_id": importApplyResultID}}, nil
}

func assertImportParticipantExecution(t *testing.T, module *Module) {
	t.Helper()
	facade, request, source := streamingApplyFixture(t, validPreviewRow, module.limits)
	facade.store = &store{}
	capabilityModule := &Module{store: facade.store, importSources: source, limits: module.limits, now: facade.now, safeDigester: facade.safeDigester}
	for _, id := range []string{ImportApplyParticipantID, IndicatorLinkParticipantID} {
		t.Run("phase capabilities/"+id, func(t *testing.T) {
			read, write, err := capabilityModule.TransactionCapabilities(id, &transactionLifecycleTx{})
			if err != nil || read.ParticipantScope() != id || write.ParticipantScope() != id {
				t.Fatalf("real owner capabilities: %v", err)
			}
			_, readImport := read.(importApplyReadCapability)
			_, readIndicator := read.(indicatorLinkReadCapability)
			_, writeImport := write.(importApplyWriteCapability)
			_, writeIndicator := write.(indicatorLinkWriteCapability)
			if readImport != (id == ImportApplyParticipantID) || writeImport != readImport || readIndicator != (id == IndicatorLinkParticipantID) || writeIndicator != readIndicator {
				t.Fatal("capability includes another participant's logical surface")
			}
			// Scope plus one owner method is the complete logical capability.
			// Extra mutation, query, transaction or connection methods are forbidden.
			if reflect.TypeOf(read).NumMethod() != 2 || reflect.TypeOf(write).NumMethod() != 2 {
				t.Fatal("physical factory exposes methods outside the selected phase")
			}
			if _, ok := read.(importApplyWriteCapability); ok {
				t.Fatal("reader exposes import mutation")
			}
			if _, ok := read.(indicatorLinkWriteCapability); ok {
				t.Fatal("reader exposes link mutation")
			}
			if _, ok := write.(importApplyReadCapability); ok {
				t.Fatal("writer exposes import validation")
			}
			if _, ok := write.(indicatorLinkReadCapability); ok {
				t.Fatal("writer exposes link validation")
			}
		})
	}
	request.IncidentID, request.ActorUserID = uuid.New(), uuid.New()
	request.ImportSessionID, request.ImportUnitID = uuid.New(), uuid.New()
	request.SourceCapability.ImportSessionID, request.SourceCapability.ImportUnitID = request.ImportSessionID, request.ImportUnitID
	request.TargetKind, request.ExtensionProfileID = targetKindNetworkFlowTable, ProfileID
	request.OwnerMappingSchemaID, request.ClientTxnID = importMappingSchemaID, "participant-import"
	participant, err := facade.ImportTransaction(request)
	if err != nil || participant.ID() != ImportApplyParticipantID {
		t.Fatalf("owner factory: %v", err)
	}
	changed := request
	changed.MappingFingerprint = strings.Repeat("b", 64)
	changedParticipant, err := facade.ImportTransaction(changed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := changedParticipant.BuildInput(context.Background(), crossownertransaction.OperationContext{}); err == nil || source.opens != 0 {
		t.Fatalf("changed mapping reached source access: %v opens=%d", err, source.opens)
	}
	// The facade must own the approved input independently of its caller's buffer.
	request.OwnerMapping[0] = '!'
	input, err := participant.BuildInput(context.Background(), crossownertransaction.OperationContext{})
	if err != nil {
		t.Fatal(err)
	}
	value, err := canonicaljson.DecodeExtension(input.CanonicalBytes)
	if err != nil || !module.ValidateTransactionInput(input.SchemaID, value) {
		t.Fatalf("real analytical producer failed current admission: %v", err)
	}
	prepared, err := participant.Prepare(context.Background(), crossownertransaction.Invocation{})
	want := []crossownertransaction.SerializationKey{
		{KeyKind: "network_flow_activity.import_unit", Key: request.ImportUnitID.String()},
		{KeyKind: "network_flow_activity.incident", Key: request.IncidentID.String()},
	}
	if err != nil || !reflect.DeepEqual(prepared.SerializationKeys, want) || source.opens != 0 {
		t.Fatalf("prepare performed source work or returned wrong keys: %+v %v opens=%d", prepared, err, source.opens)
	}
	port := &importParticipantPortFixture{}
	validated, err := participant.Validate(context.Background(), crossownertransaction.Invocation{ReadAccess: port})
	if err != nil || validated.Status != "valid" || len(validated.Findings) != 0 || !reflect.DeepEqual(port.calls, []string{"validate"}) {
		t.Fatalf("validation: %+v %v %v", validated, err, port.calls)
	}
	if _, err := participant.Write(context.Background(), crossownertransaction.Invocation{}); !errors.Is(err, crossownertransaction.ErrUnavailable) {
		t.Fatal("write admitted missing owner access")
	}
	written, err := participant.Write(context.Background(), crossownertransaction.Invocation{WriteAccess: port})
	if err != nil || written.Status != "written" || !reflect.DeepEqual(port.calls, []string{"validate", "write"}) {
		t.Fatalf("write: %+v %v %v", written, err, port.calls)
	}
	if _, ok := written.Value.(imports.ExtensionImportApplyResult); !ok {
		t.Fatal("owner receipt lost its typed boundary")
	}
}
