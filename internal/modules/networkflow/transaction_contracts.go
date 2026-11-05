package networkflow

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
)

const (
	indicatorLinkTransactionInputSchema = "cartulary.network_flow_activity.indicator_link_transaction_input.v2"
	importApplyTransactionInputSchema   = "cartulary.network_flow_activity.import_apply_transaction_input.v2"
)

// TransactionContract describes the concrete owner algorithms. Composition
// admits these identities against the generated participant contract before use.
func TransactionContract(id string) (crossownertransaction.Descriptor, bool) {
	d := crossownertransaction.Descriptor{ParticipantID: id, OwnerProfileID: ProfileID}
	switch id {
	case IndicatorLinkParticipantID:
		d.InputSchemaID = indicatorLinkTransactionInputSchema
		d.PrepareAlgorithmID = "network_flow_activity.indicator_link_prepare_v2"
		d.ValidationAlgorithmID = "network_flow_activity.indicator_link_validate_v2"
		d.WriteAlgorithmID = "network_flow_activity.indicator_link_write_v2"
		d.SerializationKeyKinds = []string{"network_flow_activity.incident", "network_flow_activity.indicator", "network_flow_activity.table"}
		d.OwnedStateFamilyIDs = []string{"network_flow_activity.indicator_bindings"}
	case ImportApplyParticipantID:
		d.InputSchemaID = importApplyTransactionInputSchema
		d.PrepareAlgorithmID = "network_flow_activity.import_apply_prepare_v2"
		d.ValidationAlgorithmID = "network_flow_activity.import_apply_validate_v2"
		d.WriteAlgorithmID = "network_flow_activity.import_apply_write_v2"
		d.SerializationKeyKinds = []string{"network_flow_activity.import_unit", "network_flow_activity.incident"}
		d.OwnedStateFamilyIDs = []string{"network_flow_activity.rejected_row_diagnostics", "network_flow_activity.rows", "network_flow_activity.tables"}
	default:
		return crossownertransaction.Descriptor{}, false
	}
	d.ContractSHA256 = crossownertransaction.DescriptorDigest(d)
	return d, true
}

// ValidateTransactionInput resolves only Network Flow's two named input
// contracts and reuses their existing owner parsers for nested request values.
// It performs no database access, authorization, or mutation.
func (m *Module) ValidateTransactionInput(schemaID string, value any) bool {
	if m == nil {
		return false
	}
	object, ok := value.(map[string]any)
	if !ok || object["schema_id"] != schemaID {
		return false
	}
	data, err := json.Marshal(object)
	if err != nil {
		return false
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(data, &raw) != nil {
		return false
	}
	switch schemaID {
	case indicatorLinkTransactionInputSchema:
		if linkMembers(raw, map[string]string{"schema_id": "string", "incident_id": "string", "actor_user_id": "string", "request": "object", "target": "object", "candidate_value": "string", "client_txn_id": "string"}) != nil {
			return false
		}
		if !transactionUUID(linkString(raw, "incident_id")) || !transactionUUID(linkString(raw, "actor_user_id")) || !canonicalIPLiteral(linkString(raw, "candidate_value")) {
			return false
		}
		client := linkString(raw, "client_txn_id")
		if client == "" || utf8.RuneCountInString(client) > 160 {
			return false
		}
		if _, failure := decodeIndicatorSelector(raw["request"], m.limits); failure != nil {
			return false
		}
		_, failure := decodeIndicatorTarget(raw["target"])
		return failure == nil
	case importApplyTransactionInputSchema:
		return m.validateImportTransactionInput(raw)
	default:
		return false
	}
}

func transactionUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func (m *Module) validateImportTransactionInput(raw map[string]json.RawMessage) bool {
	if linkMembers(raw, map[string]string{
		"schema_id": "string", "incident_id": "string", "actor_user_id": "string", "import_session_id": "string", "import_unit_id": "string",
		"source_stream_ref": "string", "source_content_sha256": "string", "mapping_fingerprint": "string",
		"owner_mapping_schema_id": "string", "owner_mapping": "object", "client_txn_id": "string",
	}) != nil {
		return false
	}
	for _, key := range []string{"incident_id", "actor_user_id", "import_session_id", "import_unit_id"} {
		if !transactionUUID(linkString(raw, key)) {
			return false
		}
	}
	stream := linkString(raw, "source_stream_ref")
	client := linkString(raw, "client_txn_id")
	if stream == "" || len(stream) > 512 || strings.ContainsRune(stream, 0) || client == "" || utf8.RuneCountInString(client) > 160 ||
		!validLowerSHA256(linkString(raw, "source_content_sha256")) || !validLowerSHA256(linkString(raw, "mapping_fingerprint")) || linkString(raw, "owner_mapping_schema_id") != importMappingSchemaID {
		return false
	}
	mapping, err := decodeApprovedMapping(raw["owner_mapping"], m.limits)
	if err != nil {
		return false
	}
	fingerprint, err := mappingFingerprint(mapping, linkString(raw, "source_content_sha256"))
	return err == nil && fingerprint == linkString(raw, "mapping_fingerprint")
}
