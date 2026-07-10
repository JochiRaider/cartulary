package networkflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
)

type NetworkFlowRowRef struct {
	NetworkFlowTableID string `json:"network_flow_table_id"`
	NetworkFlowRowID   string `json:"network_flow_row_id"`
	SourceRowNumber    int64  `json:"source_row_number"`
	MappingFingerprint string `json:"mapping_fingerprint"`
}

type IndicatorBindingRecord struct {
	BindingID                  string
	IncidentID                 uuid.UUID
	TargetIndicator            indicators.IndicatorRecord
	SelectorKind               string
	CandidateValue             string
	SourceRowRefs              []NetworkFlowRowRef
	SourceRowRefsTruncated     bool
	SourceRowRefsTotalCount    int64
	CreatedObservationRefsJSON json.RawMessage
	CreatedByUserID            uuid.UUID
	CreatedAt                  time.Time
}

type CreateIndicatorBindingParams struct {
	IncidentID              uuid.UUID
	ActorUserID             uuid.UUID
	TargetIndicator         indicators.IndicatorRecord
	SelectorKind            string
	CandidateValue          string
	SourceRowRefs           []NetworkFlowRowRef
	SourceRowRefsTruncated  bool
	SourceRowRefsTotalCount int64
	ClientTxnID             string
	RequestID               string
	CandidateDigestKeyID    string
	CandidateDigestKey      []byte
	Now                     time.Time
}

func (s *Store) CreateOrReuseIndicatorBindingTx(ctx context.Context, tx pgx.Tx, params CreateIndicatorBindingParams) (IndicatorBindingRecord, bool, error) {
	now := normalizedNow(params.Now)
	if params.IncidentID == uuid.Nil || params.ActorUserID == uuid.Nil || params.TargetIndicator.RecordID == uuid.Nil {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	if err := lockIncidentTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorBindingRecord{}, false, err
	}
	if params.TargetIndicator.IncidentID != params.IncidentID || params.TargetIndicator.DeletedAt != nil {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	if params.TargetIndicator.ValueKind != "atomic" || params.TargetIndicator.NormalizedValue == nil || *params.TargetIndicator.NormalizedValue != params.CandidateValue {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	if params.TargetIndicator.IndicatorType != "ipv4_addr" && params.TargetIndicator.IndicatorType != "ipv6_addr" {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	if !validBindingSelectorKind(params.SelectorKind) || params.CandidateValue == "" {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	if len(params.SourceRowRefs) == 0 || int64(len(params.SourceRowRefs)) > s.limits.MaxBindingSourceRowRefs || params.SourceRowRefsTotalCount < int64(len(params.SourceRowRefs)) || params.SourceRowRefsTotalCount <= 0 {
		return IndicatorBindingRecord{}, false, ErrInvalidStorageArgument
	}
	sourceRefs := append([]NetworkFlowRowRef(nil), params.SourceRowRefs...)
	sourceRefRowIDs := bindingSourceRowIDs(sourceRefs)
	sourceRefsJSON, err := json.Marshal(sourceRefs)
	if err != nil {
		return IndicatorBindingRecord{}, false, fmt.Errorf("marshal network flow binding source refs: %w", err)
	}
	for attempt := 0; attempt < 8; attempt++ {
		bindingID, err := newBindingID()
		if err != nil {
			return IndicatorBindingRecord{}, false, err
		}
		record, inserted, err := insertIndicatorBindingTx(ctx, tx, bindingID, params, sourceRefsJSON, sourceRefRowIDs, now)
		if err != nil {
			return IndicatorBindingRecord{}, false, err
		}
		if inserted {
			if err := s.insertIndicatorBindingAuditTx(ctx, tx, params, record, "network_flow_indicator_binding_created"); err != nil {
				return IndicatorBindingRecord{}, false, err
			}
			return record, false, nil
		}
		existing, found, err := getIndicatorBindingByIdentityTx(ctx, tx, params.IncidentID, params.TargetIndicator.RecordID, params.CandidateValue, sourceRefRowIDs)
		if err != nil {
			return IndicatorBindingRecord{}, false, err
		}
		if found {
			if err := s.insertIndicatorBindingAuditTx(ctx, tx, params, existing, "network_flow_indicator_binding_reused"); err != nil {
				return IndicatorBindingRecord{}, false, err
			}
			return existing, true, nil
		}
	}
	return IndicatorBindingRecord{}, false, ErrIDGenerationFailed
}

func (s *Store) GetActiveIndicator(ctx context.Context, incidentID uuid.UUID, indicatorID uuid.UUID) (indicators.IndicatorRecord, error) {
	return getActiveIndicator(ctx, s.pool, incidentID, indicatorID)
}

func getActiveIndicator(ctx context.Context, querier tableQuerier, incidentID uuid.UUID, indicatorID uuid.UUID) (indicators.IndicatorRecord, error) {
	row := querier.QueryRow(ctx, `
SELECT record_id, incident_id, indicator_type, value_kind, display_value, normalized_value,
       dedupe_key, defanged_value, hash_algorithm, hash_value, stix_pattern,
       row_version, created_at, updated_at, created_by_user_id, updated_by_user_id,
       deleted_at, deleted_by_user_id
  FROM indicators
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
`, incidentID, indicatorID)
	record, err := scanIndicatorRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return indicators.IndicatorRecord{}, ErrTableNotFound
	}
	if err != nil {
		return indicators.IndicatorRecord{}, fmt.Errorf("get active indicator for network flow binding: %w", err)
	}
	return record, nil
}

func insertIndicatorBindingTx(ctx context.Context, tx pgx.Tx, bindingID string, params CreateIndicatorBindingParams, sourceRefsJSON []byte, sourceRefRowIDs []string, now time.Time) (IndicatorBindingRecord, bool, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO network_flow_indicator_bindings (
    network_flow_indicator_binding_id, incident_id, target_indicator_record_id,
    target_indicator_type, target_indicator_value_kind, target_indicator_normalized_value,
    selector_kind, candidate_value, source_row_refs, source_row_ref_row_ids,
    source_row_refs_truncated, source_row_refs_total_count, created_by_user_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::text[], $11, $12, $13, $14
)
ON CONFLICT DO NOTHING
RETURNING network_flow_indicator_binding_id, incident_id, target_indicator_record_id,
          target_indicator_type, target_indicator_value_kind, target_indicator_normalized_value,
          selector_kind, candidate_value, source_row_refs, source_row_refs_truncated,
          source_row_refs_total_count, created_by_user_id, created_at
`, bindingID, params.IncidentID, params.TargetIndicator.RecordID, params.TargetIndicator.IndicatorType, params.TargetIndicator.ValueKind, *params.TargetIndicator.NormalizedValue, params.SelectorKind, params.CandidateValue, string(sourceRefsJSON), sourceRefRowIDs, params.SourceRowRefsTruncated, params.SourceRowRefsTotalCount, params.ActorUserID, now)
	record, err := scanIndicatorBinding(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorBindingRecord{}, false, nil
	}
	if err != nil {
		return IndicatorBindingRecord{}, false, fmt.Errorf("insert network flow indicator binding: %w", err)
	}
	return record, true, nil
}

func getIndicatorBindingByIdentityTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, indicatorID uuid.UUID, candidateValue string, sourceRefRowIDs []string) (IndicatorBindingRecord, bool, error) {
	row := tx.QueryRow(ctx, `
SELECT network_flow_indicator_binding_id, incident_id, target_indicator_record_id,
       target_indicator_type, target_indicator_value_kind, target_indicator_normalized_value,
       selector_kind, candidate_value, source_row_refs, source_row_refs_truncated,
       source_row_refs_total_count, created_by_user_id, created_at
  FROM network_flow_indicator_bindings
 WHERE incident_id = $1
   AND target_indicator_record_id = $2
   AND candidate_value = $3
   AND source_row_ref_row_ids = $4::text[]
`, incidentID, indicatorID, candidateValue, sourceRefRowIDs)
	record, err := scanIndicatorBinding(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorBindingRecord{}, false, nil
	}
	if err != nil {
		return IndicatorBindingRecord{}, false, fmt.Errorf("get network flow indicator binding by identity: %w", err)
	}
	return record, true, nil
}

func (s *Store) insertIndicatorBindingAuditTx(ctx context.Context, tx pgx.Tx, params CreateIndicatorBindingParams, record IndicatorBindingRecord, eventKind string) error {
	candidateDigest, keyID := SafeDigest(params.CandidateDigestKeyID, params.CandidateDigestKey, "indicator_candidate_value", record.CandidateValue)
	return insertNetworkFlowAuditEvent(ctx, tx, networkFlowAuditEvent{
		ActorUserID: &params.ActorUserID,
		IncidentID:  &params.IncidentID,
		EventKind:   eventKind,
		ClientTxnID: optionalStringPtr(params.ClientTxnID),
		RequestID:   optionalStringPtr(params.RequestID),
		AfterJSON: map[string]any{
			"incident_id":                       params.IncidentID.String(),
			"actor_user_id":                     params.ActorUserID.String(),
			"network_flow_indicator_binding_id": record.BindingID,
			"target_indicator_record_id":        record.TargetIndicator.RecordID.String(),
			"selector_kind":                     record.SelectorKind,
			"candidate_value_digest":            candidateDigest,
			"candidate_value_digest_key_id":     keyID,
			"source_row_ref_count":              len(record.SourceRowRefs),
			"network_flow.audit_event_code":     eventKind,
			"network_flow.audit_resource_id":    record.BindingID,
		},
	})
}

func scanIndicatorBinding(row pgx.Row) (IndicatorBindingRecord, error) {
	var record IndicatorBindingRecord
	var indicatorID uuid.UUID
	var indicatorNormalizedValue string
	var sourceRefsJSON []byte
	if err := row.Scan(
		&record.BindingID,
		&record.IncidentID,
		&indicatorID,
		&record.TargetIndicator.IndicatorType,
		&record.TargetIndicator.ValueKind,
		&indicatorNormalizedValue,
		&record.SelectorKind,
		&record.CandidateValue,
		&sourceRefsJSON,
		&record.SourceRowRefsTruncated,
		&record.SourceRowRefsTotalCount,
		&record.CreatedByUserID,
		&record.CreatedAt,
	); err != nil {
		return IndicatorBindingRecord{}, err
	}
	if err := json.Unmarshal(sourceRefsJSON, &record.SourceRowRefs); err != nil {
		return IndicatorBindingRecord{}, fmt.Errorf("decode network flow indicator binding source refs: %w", err)
	}
	record.TargetIndicator.RecordID = indicatorID
	record.TargetIndicator.IncidentID = record.IncidentID
	record.TargetIndicator.NormalizedValue = &indicatorNormalizedValue
	record.CreatedAt = record.CreatedAt.UTC()
	record.CreatedObservationRefsJSON = json.RawMessage(`[]`)
	return record, nil
}

func scanIndicatorRecord(row pgx.Row) (indicators.IndicatorRecord, error) {
	var record indicators.IndicatorRecord
	var deletedAt pgtype.Timestamptz
	var deletedBy pgtype.UUID
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&record.NormalizedValue,
		&record.DedupeKey,
		&record.DefangedValue,
		&record.HashAlgorithm,
		&record.HashValue,
		&record.STIXPattern,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
		&deletedAt,
		&deletedBy,
	); err != nil {
		return indicators.IndicatorRecord{}, err
	}
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	if deletedBy.Valid {
		value, err := uuid.FromBytes(deletedBy.Bytes[:])
		if err != nil {
			return indicators.IndicatorRecord{}, fmt.Errorf("decode deleted indicator user id: %w", err)
		}
		record.DeletedByUserID = &value
	}
	return record, nil
}

func bindingSourceRowIDs(refs []NetworkFlowRowRef) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref.NetworkFlowRowID)
	}
	sort.Strings(out)
	return out
}

func validBindingSelectorKind(kind string) bool {
	switch kind {
	case "row_field_value", "row_refs", "graph_vertex", "graph_edge":
		return true
	default:
		return false
	}
}

func newBindingID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate network flow indicator binding id: %w", err)
	}
	return "nfb_" + hex.EncodeToString(bytes[:]), nil
}
