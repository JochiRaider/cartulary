package networkflow

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const PortabilityTransactionParticipantID = "network_flow_activity.incident_import_v1"
const portabilityParticipantID = "network_flow_activity.incident_portability_v1"
const portabilityInputSchema = "cartulary.network_flow.incident_import_input.v1"
const portabilitySpecializationSchema = "cartulary.extension_participant_specialization.v5"

type portabilityParticipant struct{ limits EffectiveLimits }

func (m *Module) IncidentPortability() incidentbundles.ExtensionParticipant {
	if m == nil {
		return nil
	}
	return &portabilityParticipant{limits: m.limits}
}
func (*portabilityParticipant) ID() string { return portabilityParticipantID }
func (*portabilityParticipant) SpecializationSchemaID() string {
	return portabilitySpecializationSchema
}
func (*portabilityParticipant) ContractSHA256() string { return portabilityContractSHA256 }
func (*portabilityParticipant) Export(ctx context.Context, invocation incidentbundles.ExportInvocation) (incidentbundles.ExportResult, error) {
	access, err := requirePortabilityExportAccess(invocation)
	if err != nil {
		return incidentbundles.ExportResult{}, err
	}
	if err := access.ExportRetained(ctx, invocation.Members, invocation.Actors); err != nil {
		return incidentbundles.ExportResult{}, err
	}
	return incidentbundles.ExportResult{SchemaID: incidentbundles.ExtensionExportResultSchema, Kind: "members", ContentSchemaID: portabilityContentSchema, ContractMajor: 10, StateVersion: 7}, nil
}
func (p *portabilityParticipant) PrepareImport(ctx context.Context, invocation incidentbundles.ImportInvocation, scope *incidentbundles.PortabilityStagedOutputScope) (incidentbundles.ImportPreparation, error) {
	prepared, err := prepareNetworkFlow(ctx, invocation, p.limits)
	if err != nil {
		return incidentbundles.ImportPreparation{}, err
	}
	graphMembers := []incidentportability.Member{}
	for _, graph := range prepared.graphs {
		if graph.members != nil {
			graphMembers = append(graphMembers, graph.members.Members()...)
		}
	}
	input, err := canonicaljson.MarshalExtension(map[string]any{"schema_id": portabilityInputSchema, "operation_id": prepared.operationID, "incident_id": prepared.incidentID.String(), "actor_user_id": prepared.actorID.String(), "catalog_sha256": invocation.CatalogSHA256, "members": prepared.members.Members(), "prepared_graphs": graphMembers, "counts": prepared.counts})
	if err != nil {
		return incidentbundles.ImportPreparation{}, err
	}
	participant := &portabilityImportTransaction{prepared: prepared, input: input}
	return incidentbundles.ImportPreparation{SchemaID: incidentbundles.ExtensionImportResultSchema, Status: "prepared", ParticipantInput: input, ParticipantInputSHA256: sha256Hex(input), StagedOutputRefs: []string{}, TransactionParticipant: participant}, nil
}

type portabilityImportTransaction struct {
	prepared *preparedNetworkFlow
	input    []byte
}

func (*portabilityImportTransaction) ID() string { return PortabilityTransactionParticipantID }
func (p *portabilityImportTransaction) BuildInput(_ context.Context, operation crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	if p == nil || p.prepared == nil || operation.OperationID != p.prepared.operationID {
		return crossownertransaction.Input{}, errPortableNetworkFlow
	}
	return crossownertransaction.Input{SchemaID: portabilityInputSchema, CanonicalBytes: append([]byte(nil), p.input...)}, nil
}
func (p *portabilityImportTransaction) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{{KeyKind: "network_flow_activity.incident", Key: p.prepared.incidentID.String()}}}, nil
}
func (p *portabilityImportTransaction) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	access, ok := invocation.ReadAccess.(interface {
		ValidatePortableState(context.Context, *preparedNetworkFlow) error
	})
	if !ok {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := access.ValidatePortableState(ctx, p.prepared); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}
func (p *portabilityImportTransaction) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	access, ok := invocation.WriteAccess.(interface {
		WritePortableState(context.Context, *preparedNetworkFlow) error
	})
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	if err := access.WritePortableState(ctx, p.prepared); err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(nil), nil
}

func validatePortabilityTransactionInput(value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 8 {
		return false
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return false
	}
	var input struct {
		SchemaID    string                       `json:"schema_id"`
		OperationID string                       `json:"operation_id"`
		IncidentID  string                       `json:"incident_id"`
		ActorID     string                       `json:"actor_user_id"`
		Catalog     string                       `json:"catalog_sha256"`
		Members     []incidentportability.Member `json:"members"`
		Graphs      []incidentportability.Member `json:"prepared_graphs"`
		Counts      map[string]int64             `json:"counts"`
	}
	if _, err := decodeClosedImportObject(raw, "schema_id operation_id incident_id actor_user_id catalog_sha256 members prepared_graphs counts", "", "", &input); err != nil {
		return false
	}
	if input.SchemaID != portabilityInputSchema || !transactionUUID(input.OperationID) || !transactionUUID(input.IncidentID) || !transactionUUID(input.ActorID) || !validLowerSHA256(input.Catalog) || len(input.Members) != 5 || len(input.Counts) != 5 {
		return false
	}
	for _, name := range []string{"tables", "observations", "diagnostics", "indicator_bindings", "graph_views"} {
		n, ok := input.Counts[name]
		if !ok || n < 0 {
			return false
		}
	}
	// The prepared descriptor binds only the five owner members and exact
	// reconstructed graph objects. Nested unknown fields cannot be dropped.
	closed, err := json.Marshal(input)
	if err != nil || !canonicalEqual(json.RawMessage(raw), json.RawMessage(closed)) {
		return false
	}
	expected := map[string]string{}
	for _, name := range portabilityFamilyPaths {
		expected[portabilityMemberPrefix+name] = strings.TrimSuffix(name, ".ndjson")
	}
	previous := ""
	for _, member := range input.Members {
		family, ok := expected[member.Path]
		if !ok || member.Path <= previous || member.ByteSize < 0 || !validLowerSHA256(member.SHA256) || (input.Counts[family] == 0) != (member.ByteSize == 0) {
			return false
		}
		if member.ByteSize == 0 && member.SHA256 != sha256Hex(nil) {
			return false
		}
		previous = member.Path
	}
	if int64(len(input.Graphs)) > input.Counts["graph_views"] {
		return false
	}
	previous = ""
	const prefix = "prepared/network_flow_activity/graphs/"
	for _, member := range input.Graphs {
		if !strings.HasPrefix(member.Path, prefix) || !strings.HasSuffix(member.Path, ".json") || member.Path <= previous || member.ByteSize <= 0 || !validLowerSHA256(member.SHA256) {
			return false
		}
		id := strings.TrimSuffix(strings.TrimPrefix(member.Path, prefix), ".json")
		if !graphViewIDPattern.MatchString(id) {
			return false
		}
		previous = member.Path
	}
	return true
}

const portabilityContractSHA256 = "6f78f64d787c08fd00f90a304e4f26e67a9a94fc4aea441d30fa9a214a65515c"
