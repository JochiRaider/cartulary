package indicators

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

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

func normalizedIndicatorCreateHash(input indicatorUpsertInput) []byte {
	payload := map[string]any{
		"indicator.indicator_type": input.IndicatorType,
		"indicator.value_kind":     input.ValueKind,
		"indicator.display_value":  input.DisplayValue,
	}
	for key, value := range map[string]*string{
		"indicator.normalized_value": input.NormalizedValue,
		"indicator.defanged_value":   input.DefangedValue,
		"indicator.hash_algorithm":   input.HashAlgorithm,
		"indicator.hash_value":       input.HashValue,
		"indicator.stix_pattern":     input.STIXPattern,
	} {
		if value != nil {
			payload[key] = *value
		}
	}
	return hashReplayRequest(marshalReplayRequest(payload))
}
