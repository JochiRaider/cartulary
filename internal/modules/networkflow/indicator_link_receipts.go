package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/google/uuid"
)

var errIndicatorLinkReceiptIncompatible = errors.New("indicator link retained receipt is incompatible")

func validateIndicatorLinkReceipt(payload map[string]any, status int, incidentID uuid.UUID) error {
	duplicate, ok := payload["duplicate"].(bool)
	if !ok || len(payload) != 3 || payload["schema_id"] != schemaIndicatorLinkResult || (duplicate && status != 200) || (!duplicate && status != 201) {
		return errIndicatorLinkReceiptIncompatible
	}
	raw, err := json.Marshal(payload["binding"])
	if err != nil {
		return errIndicatorLinkReceiptIncompatible
	}
	return validateIndicatorBindingBytes(raw, incidentID)
}

func validateIndicatorBindingBytes(raw []byte, incidentID uuid.UUID) error {
	object, apiErr := linkObject(raw, "binding")
	if apiErr != nil {
		return errIndicatorLinkReceiptIncompatible
	}
	if linkMembers(object, map[string]string{
		"network_flow_indicator_binding_id": "string", "incident_id": "string", "target_indicator_ref": "object", "selector_kind": "string", "candidate_value": "string",
		"source_row_refs": "array", "source_row_refs_truncated": "boolean", "source_row_refs_total_count": "number", "created_observation_refs": "array", "created_by_user_id": "string", "created_at": "string",
	}) != nil {
		return errIndicatorLinkReceiptIncompatible
	}
	var binding struct {
		BindingID      string            `json:"network_flow_indicator_binding_id"`
		IncidentID     uuid.UUID         `json:"incident_id"`
		SelectorKind   string            `json:"selector_kind"`
		CandidateValue string            `json:"candidate_value"`
		Refs           []json.RawMessage `json:"source_row_refs"`
		Truncated      bool              `json:"source_row_refs_truncated"`
		Total          int64             `json:"source_row_refs_total_count"`
		Observations   []any             `json:"created_observation_refs"`
		Actor          uuid.UUID         `json:"created_by_user_id"`
		CreatedAt      string            `json:"created_at"`
	}
	if json.Unmarshal(raw, &binding) != nil || !linkBindingIDPattern.MatchString(binding.BindingID) || !linkUUIDPattern.MatchString(linkString(object, "incident_id")) || !linkUUIDPattern.MatchString(linkString(object, "created_by_user_id")) || binding.IncidentID == uuid.Nil || binding.IncidentID != incidentID || binding.Actor == uuid.Nil || !validBindingSelectorKind(binding.SelectorKind) {
		return errIndicatorLinkReceiptIncompatible
	}
	kind, ok := indicators.CanonicalIPIndicatorType(binding.CandidateValue)
	if !ok {
		return errIndicatorLinkReceiptIncompatible
	}
	target, apiErr := linkObject(object["target_indicator_ref"], "target_indicator_ref")
	if apiErr != nil || linkMembers(target, map[string]string{"indicator_id": "string", "indicator_type": "string", "value_kind": "string", "normalized_value": "string"}) != nil {
		return errIndicatorLinkReceiptIncompatible
	}
	id, err := uuid.Parse(linkString(target, "indicator_id"))
	if err != nil || id == uuid.Nil || !linkUUIDPattern.MatchString(linkString(target, "indicator_id")) || linkString(target, "indicator_type") != kind || linkString(target, "value_kind") != "atomic" || linkString(target, "normalized_value") != binding.CandidateValue {
		return errIndicatorLinkReceiptIncompatible
	}
	created, err := time.Parse(time.RFC3339Nano, binding.CreatedAt)
	if err != nil || timestamp(created) != binding.CreatedAt || len(binding.Observations) != 0 || len(binding.Refs) < 1 || len(binding.Refs) > 1000 || binding.Total < int64(len(binding.Refs)) || binding.Truncated != (binding.Total > int64(len(binding.Refs))) {
		return errIndicatorLinkReceiptIncompatible
	}
	if binding.SelectorKind == "row_field_value" && (len(binding.Refs) != 1 || binding.Truncated) {
		return errIndicatorLinkReceiptIncompatible
	}
	if binding.SelectorKind == "row_refs" && binding.Truncated {
		return errIndicatorLinkReceiptIncompatible
	}
	seen := map[string]bool{}
	for _, rawRef := range binding.Refs {
		ref, apiErr := decodeLinkRowRef(rawRef)
		if apiErr != nil || seen[ref.NetworkFlowRowID] {
			return errIndicatorLinkReceiptIncompatible
		}
		seen[ref.NetworkFlowRowID] = true
	}
	return nil
}

// Admission validates retained bytes in the existing read-only state snapshot.
// It never translates receipts: older result identifiers cannot be replayed by
// this coordinated server/browser contract correction.
func validatePersistedIndicatorLinkFamily(ctx context.Context, reader extensionstore.Querier) error {
	rows, err := reader.Query(ctx, `SELECT network_flow_indicator_binding_id, incident_id, target_indicator_record_id,
		target_indicator_type, target_indicator_value_kind, target_indicator_normalized_value,
		selector_kind, candidate_value, source_row_refs, source_row_refs_truncated,
		source_row_refs_total_count, created_by_user_id, created_at
		FROM network_flow_indicator_bindings ORDER BY network_flow_indicator_binding_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		binding, err := scanIndicatorBinding(rows)
		if err != nil {
			rows.Close()
			return err
		}
		raw, err := json.Marshal(indicatorBindingResource(binding))
		if err != nil || validateIndicatorBindingBytes(raw, binding.IncidentID) != nil {
			rows.Close()
			return errIndicatorLinkReceiptIncompatible
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	after := uuid.Nil
	for {
		records, next, err := authn.ReadRouteIdempotencyPage(ctx, reader, []string{routeKeyIndicatorLinksCreate}, after)
		if err != nil {
			return err
		}
		for _, record := range records {
			incidentID, err := uuid.Parse(record.ScopeKey)
			if err != nil || record.ActorUserID == uuid.Nil || record.ClientTxnID == "" || len(record.RequestHash) != sha256.Size {
				return errIndicatorLinkReceiptIncompatible
			}
			payload, err := decodeStoredNetworkFlowResponse(record.ResponseJSON)
			if err != nil || validateIndicatorLinkReceipt(payload, record.StatusCode, incidentID) != nil {
				return errIndicatorLinkReceiptIncompatible
			}
		}
		if len(records) < 128 {
			return nil
		}
		after = next
	}
}

type indicatorLinkReceiptAdapter struct {
	reader *authn.Store
	store  *store
}

func (s *indicatorLinkReceiptAdapter) replay(ctx context.Context, key authn.RouteIdempotencyKey, requestHash []byte, incidentID uuid.UUID) (indicatorLinkOutcome, bool, *semanticFailure) {
	existing, err := s.reader.GetRouteIdempotency(ctx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return indicatorLinkOutcome{}, true, clientTxnFailure(key.ClientTxnID)
		}
		payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
		if err != nil {
			return indicatorLinkOutcome{}, true, internalSemanticFailure(err)
		}
		if err := validateIndicatorLinkReceipt(payload, existing.StatusCode, incidentID); err != nil {
			return indicatorLinkOutcome{}, true, internalSemanticFailure(err)
		}
		indicatorID, err := indicatorIDFromLinkPayload(payload)
		if err != nil {
			return indicatorLinkOutcome{}, true, internalSemanticFailure(err)
		}
		if _, err := s.store.GetActiveIndicator(ctx, incidentID, indicatorID); err != nil {
			if errors.Is(err, errTableNotFound) {
				return indicatorLinkOutcome{}, true, indicatorLinkForbidden("", "", "existing_indicator", "target_not_visible")
			}
			return indicatorLinkOutcome{}, true, internalSemanticFailure(err)
		}
		return indicatorLinkOutcome{receipt: &storedMutationReceipt{payload: payload, status: existing.StatusCode}}, true, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return indicatorLinkOutcome{}, false, internalSemanticFailure(err)
	}
	return indicatorLinkOutcome{}, false, nil
}
func (indicatorLinkReceiptAdapter) saveTx(ctx context.Context, tx pgx.Tx, mutation indicatorLinkMutation, binding indicatorBindingRecord, duplicate bool) error {
	payload, status := indicatorLinkOutcomeResponse(indicatorLinkOutcome{binding: binding, duplicate: duplicate})
	return authn.InsertRouteIdempotencyPayload(ctx, tx, indicatorLinkIdempotencyKey(mutation.Actor.ID, mutation.IncidentID, mutation.Request.ClientTxnID), nil, mutation.RequestHash, status, payload)
}
func indicatorLinkOutcomeResponse(outcome indicatorLinkOutcome) (map[string]any, int) {
	if outcome.receipt != nil {
		return outcome.receipt.payload, outcome.receipt.status
	}
	status := 201
	if outcome.duplicate {
		status = 200
	}
	return indicatorLinkPayload(outcome.binding, outcome.duplicate), status
}
