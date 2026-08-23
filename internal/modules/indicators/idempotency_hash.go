package indicators

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func createIndicatorRequestHash(command CreateCommand) []byte {
	return hashReplayRequest(createIndicatorRequestPreimage(command))
}

func createIndicatorRequestPreimage(command CreateCommand) []byte {
	payload := map[string]any{
		"view_schema_id":           ViewSchemaID,
		"client_txn_id":            command.ClientTxnID,
		"indicator.indicator_type": command.IndicatorType,
		"indicator.value_kind":     command.ValueKind,
		"indicator.display_value":  command.DisplayValue,
	}
	for key, value := range map[string]*string{
		"indicator.normalized_value": command.NormalizedValue,
		"indicator.defanged_value":   command.DefangedValue,
		"indicator.hash_algorithm":   command.HashAlgorithm,
		"indicator.hash_value":       command.HashValue,
		"indicator.stix_pattern":     command.STIXPattern,
	} {
		if value != nil {
			payload[key] = *value
		}
	}
	return marshalReplayRequest(payload)
}

func observationCreateRequestHash(params IndicatorObservationCreateParams) []byte {
	return hashReplayRequest(observationCreateRequestPreimage(params))
}

func observationCreateRequestPreimage(params IndicatorObservationCreateParams) []byte {
	return marshalReplayRequest(struct {
		ClientTxnID               string     `json:"client_txn_id"`
		BaseRowVersion            int64      `json:"base_row_version"`
		SourceFieldKey            string     `json:"source_field_key"`
		SpanStartByte             int        `json:"span_start_byte"`
		SpanEndByte               int        `json:"span_end_byte"`
		ParsedIndicatorType       *string    `json:"parsed_indicator_type,omitempty"`
		ResolvedIndicatorRecordID *uuid.UUID `json:"resolved_indicator_record_id,omitempty"`
	}{
		ClientTxnID: params.ClientTxnID, BaseRowVersion: params.BaseRowVersion,
		SourceFieldKey: params.SourceFieldKey, SpanStartByte: params.SpanStartByte,
		SpanEndByte: params.SpanEndByte, ParsedIndicatorType: params.ParsedIndicatorType,
		ResolvedIndicatorRecordID: params.ResolvedIndicatorRecordID,
	})
}

func observationResolveRequestHash(params IndicatorObservationResolveParams) []byte {
	return hashReplayRequest(observationResolveRequestPreimage(params))
}

func observationResolveRequestPreimage(params IndicatorObservationResolveParams) []byte {
	return marshalReplayRequest(struct {
		ClientTxnID               string    `json:"client_txn_id"`
		BaseRowVersion            int64     `json:"base_row_version"`
		ResolvedIndicatorRecordID uuid.UUID `json:"resolved_indicator_record_id"`
	}{params.ClientTxnID, params.BaseRowVersion, params.ResolvedIndicatorRecordID})
}

func observationActionRequestHash(params IndicatorObservationActionParams) []byte {
	return hashReplayRequest(observationActionRequestPreimage(params))
}

func observationActionRequestPreimage(params IndicatorObservationActionParams) []byte {
	return marshalReplayRequest(struct {
		ClientTxnID    string `json:"client_txn_id"`
		BaseRowVersion int64  `json:"base_row_version"`
	}{params.ClientTxnID, params.BaseRowVersion})
}

func lifecycleAppendRequestHash(params IndicatorLifecycleAppendParams) []byte {
	return hashReplayRequest(lifecycleAppendRequestPreimage(params))
}

func lifecycleAppendRequestPreimage(params IndicatorLifecycleAppendParams) []byte {
	return marshalReplayRequest(struct {
		ClientTxnID    string      `json:"client_txn_id"`
		BaseRowVersion int64       `json:"base_row_version"`
		LifecycleState string      `json:"lifecycle_state"`
		ValidFrom      time.Time   `json:"valid_from"`
		ValidTo        *time.Time  `json:"valid_to"`
		Confidence     *int        `json:"confidence"`
		Rationale      *string     `json:"rationale"`
		SupportRefs    []uuid.UUID `json:"support_refs"`
		Assessor       *string     `json:"assessor"`
	}{
		params.ClientTxnID, params.BaseRowVersion, params.LifecycleState,
		params.ValidFrom, params.ValidTo, params.Confidence, params.Rationale,
		params.SupportRefs, params.Assessor,
	})
}

func marshalReplayRequest(value any) []byte {
	preimage, _ := json.Marshal(value)
	return preimage
}

func hashReplayRequest(preimage []byte) []byte {
	digest := sha256.Sum256(preimage)
	return append([]byte(nil), digest[:]...)
}
