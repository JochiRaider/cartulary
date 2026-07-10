package rollbackprovider

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type MentionProvider struct{}

var _ rollbackcontract.NonRowTargetProvider = MentionProvider{}

func NewMentionProvider() MentionProvider { return MentionProvider{} }

type mentionIdentity struct {
	mentionID      uuid.UUID
	sourceRecordID uuid.UUID
	sourceFieldKey string
	entityType     string
	status         string
	rowVersion     int64
}

func (MentionProvider) DescribeTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	identity, err := parseMentionTarget(request.Target)
	if err != nil {
		return rollbackcontract.TargetDescriptor{}, err
	}
	descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: []uuid.UUID{identity.sourceRecordID}}
	var currentSourceID uuid.UUID
	var incidentID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT em.source_record_id, r.incident_id
  FROM entity_mentions em
  JOIN records r ON r.record_id = em.source_record_id
 WHERE em.entity_mention_id = $1
`, identity.mentionID).Scan(&currentSourceID, &incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return descriptor, rollbackcontract.ErrTargetNotFound
		}
		return descriptor, err
	}
	if incidentID != request.Target.IncidentID || currentSourceID != identity.sourceRecordID {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	return descriptor, nil
}

func (p MentionProvider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	descriptor, err := p.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: request.Target})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	identity, err := parseMentionTarget(request.Target)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	before, err := loadMentionValueTx(ctx, tx, identity.mentionID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	if err := validateResolvedTargetTx(ctx, tx, request.Target.IncidentID, identity, request.Target.BeforeValue); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	retained := request.Target.BeforeValue
	resolvedRecordID, _, err := mentionNullableUUID(retained, "resolved_record_id")
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, rollbackcontract.ErrTargetNotReversible
	}
	resolvedByUserID, _, err := mentionNullableUUID(retained, "resolved_by_user_id")
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, rollbackcontract.ErrTargetNotReversible
	}
	resolvedAt, _, err := mentionNullableTime(retained, "resolved_at")
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, rollbackcontract.ErrTargetNotReversible
	}
	resolutionMethod := nullableMentionText(retained, "resolution_method")
	tag, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET source_record_id = $2,
       entity_type = $3,
       source_field_key = $4,
       origin_kind = $5,
       origin_locator = $6,
       raw_text = $7,
       normalized_text = $8,
       resolution_status = $9,
       row_version = row_version + 1,
       resolved_record_id = $10,
       resolved_by_user_id = $11,
       resolved_at = $12,
       resolution_method = $13
 WHERE entity_mention_id = $1
`, identity.mentionID, identity.sourceRecordID, identity.entityType, identity.sourceFieldKey,
		requiredMentionText(retained, "origin_kind"), requiredMentionText(retained, "origin_locator"),
		requiredMentionText(retained, "raw_text"), requiredMentionText(retained, "normalized_text"), identity.status,
		resolvedRecordID, resolvedByUserID, resolvedAt, resolutionMethod)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ApplyInverseResult{}, rollbackcontract.ErrTargetNotFound
	}
	after, err := loadMentionValueTx(ctx, tx, identity.mentionID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	var sourceType string
	if err := tx.QueryRow(ctx, `SELECT record_type FROM records WHERE record_id = $1`, identity.sourceRecordID).Scan(&sourceType); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	keys := mentionChangedFieldKeys(sourceType, identity.sourceFieldKey)
	return rollbackcontract.ApplyInverseResult{
		AffectedRecordIDs: descriptor.AffectedRecordIDs,
		BeforeValue:       before,
		AfterValue:        after,
		ChangedFieldKeys:  map[uuid.UUID][]string{identity.sourceRecordID: keys},
	}, nil
}

func mentionChangedFieldKeys(sourceType string, sourceFieldKey string) []string {
	keys := []string{sourceFieldKey}
	if sourceType == "timeline_event" {
		keys = append(keys, "timeline.has_unresolved_mentions")
	}
	sort.Strings(keys)
	return keys
}

func parseMentionTarget(target rollbackcontract.NonRowTarget) (mentionIdentity, error) {
	if target.TargetKind != "entity_mention" || target.OperationKind != "patch" || target.BeforeValue == nil {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	mentionID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotFound
	}
	retainedMentionID, err := requiredMentionUUID(target.BeforeValue, "entity_mention_id")
	if err != nil || retainedMentionID != mentionID {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	sourceID, err := requiredMentionUUID(target.BeforeValue, "source_record_id")
	if err != nil {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	entityType := requiredMentionText(target.BeforeValue, "entity_type")
	if entityType != "host" && entityType != "identity" {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	status := requiredMentionText(target.BeforeValue, "resolution_status")
	if status != "unresolved" && status != "resolved" && status != "dismissed" {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	fieldKey := requiredMentionText(target.BeforeValue, "source_field_key")
	rowVersion, valid := mentionInt64(target.BeforeValue["row_version"])
	if fieldKey == "" || !valid || rowVersion <= 0 {
		return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	for _, key := range []string{"origin_kind", "origin_locator", "raw_text", "normalized_text"} {
		if _, valid := target.BeforeValue[key].(string); !valid {
			return mentionIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
	}
	return mentionIdentity{mentionID: mentionID, sourceRecordID: sourceID, sourceFieldKey: fieldKey, entityType: entityType, status: status, rowVersion: rowVersion}, nil
}

func validateResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, identity mentionIdentity, value map[string]any) error {
	resolvedID, _, err := mentionNullableUUID(value, "resolved_record_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	resolvedBy, _, err := mentionNullableUUID(value, "resolved_by_user_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	resolvedAt, _, err := mentionNullableTime(value, "resolved_at")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	if identity.status == "resolved" {
		if resolvedID == nil || resolvedBy == nil || resolvedAt == nil {
			return rollbackcontract.ErrTargetNotReversible
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $2 AND record_type = $3)`, incidentID, *resolvedID, identity.entityType).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return rollbackcontract.ErrTargetNotReversible
		}
	} else if resolvedID != nil || resolvedBy != nil || resolvedAt != nil || nullableMentionText(value, "resolution_method") != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func loadMentionValueTx(ctx context.Context, tx pgx.Tx, mentionID uuid.UUID) (map[string]any, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'entity_mention_id', entity_mention_id::text,
    'source_record_id', source_record_id::text,
    'entity_type', entity_type,
    'source_field_key', source_field_key,
    'origin_kind', origin_kind,
    'origin_locator', origin_locator,
    'raw_text', raw_text,
    'normalized_text', normalized_text,
    'resolution_status', resolution_status,
    'row_version', row_version,
    'ordinal', ordinal,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'resolved_record_id', resolved_record_id::text,
    'resolved_by_user_id', resolved_by_user_id::text,
    'resolved_at', resolved_at,
    'resolution_method', resolution_method
)
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rollbackcontract.ErrTargetNotFound
		}
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func requiredMentionUUID(value map[string]any, key string) (uuid.UUID, error) {
	raw, valid := value[key].(string)
	if !valid || strings.TrimSpace(raw) == "" {
		return uuid.Nil, errors.New("missing uuid")
	}
	return uuid.Parse(raw)
}

func mentionNullableUUID(value map[string]any, key string) (*uuid.UUID, bool, error) {
	raw, present := value[key]
	if !present || raw == nil || raw == "" {
		return nil, present, nil
	}
	text, valid := raw.(string)
	if !valid {
		return nil, true, errors.New("invalid uuid")
	}
	parsed, err := uuid.Parse(text)
	return &parsed, true, err
}

func mentionNullableTime(value map[string]any, key string) (*time.Time, bool, error) {
	raw, present := value[key]
	if !present || raw == nil || raw == "" {
		return nil, present, nil
	}
	text, valid := raw.(string)
	if !valid {
		return nil, true, errors.New("invalid timestamp")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return nil, true, err
	}
	utc := parsed.UTC()
	return &utc, true, nil
}

func requiredMentionText(value map[string]any, key string) string {
	text, _ := value[key].(string)
	return text
}

func nullableMentionText(value map[string]any, key string) *string {
	text, valid := value[key].(string)
	if !valid || strings.TrimSpace(text) == "" {
		return nil
	}
	return &text
}

func mentionInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case float64:
		return int64(typed), typed == float64(int64(typed))
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}
