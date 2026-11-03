package extensions

import (
	"encoding/json"
	"errors"
)

// TransactionLimits is a value projection of the admitted current protocol.
// Application composition supplies it to the shared transaction coordinator.
type TransactionLimits struct {
	MinimumParticipants             int   `json:"minimum_participants"`
	Participants                    int   `json:"maximum_participants"`
	ParticipantInputBytes           int64 `json:"per_participant_input_max_bytes"`
	AggregateInputBytes             int64 `json:"aggregate_input_max_bytes"`
	PrepareResultBytes              int64 `json:"per_participant_prepare_result_max_bytes"`
	AggregatePrepareResultBytes     int64 `json:"aggregate_prepare_result_max_bytes"`
	SerializationKeysPerParticipant int   `json:"serialization_keys_per_participant"`
	AggregateSerializationKeys      int   `json:"aggregate_serialization_keys"`
	ResultBytes                     int64 `json:"result_max_bytes"`
	ValidationFindings              int   `json:"validation_findings"`
}

func (c *Coordinator) TransactionLimits() TransactionLimits { return c.transactionLimits }

func readTransactionLimits(source ArtifactSource) (TransactionLimits, error) {
	object, _, err := readArtifactObject(source, "contracts/extensions/specification/shared-protocols.json")
	if err != nil {
		return TransactionLimits{}, err
	}
	protocol, ok := object["transaction"].(map[string]any)
	if object["schema_id"] != "cartulary.extension_shared_protocol_contract_set.v5" || !ok || protocol["canonical_encoding_id"] != "extension_registry_canonical_json_v1" || protocol["input_byte_admission"] != "require_exact_canonical_bytes" || protocol["prepare_result_schema_id"] != "cartulary.extension_transaction_participant_prepare_result.v2" || protocol["validation_result_schema_id"] != "cartulary.extension_transaction_participant_validation_result.v2" || protocol["write_result_schema_id"] != "cartulary.extension_transaction_participant_write_result.v2" {
		return TransactionLimits{}, invalidArtifact("shared_transaction_protocol", errors.New("unsupported transaction byte contract"))
	}
	// Package admission has already checked the exact generated schema, including
	// constants derived from the same authored protocol as binding declarations.
	data, err := json.Marshal(object["transaction"])
	if err != nil {
		return TransactionLimits{}, invalidArtifact("shared_transaction_protocol", err)
	}
	var limits TransactionLimits
	if err := json.Unmarshal(data, &limits); err != nil {
		return TransactionLimits{}, invalidArtifact("shared_transaction_protocol", err)
	}
	return limits, nil
}
