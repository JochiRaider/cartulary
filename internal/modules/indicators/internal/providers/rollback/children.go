package rollback

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

type ChildProvider struct{}

var _ rollbackcontract.NonRowTargetProvider = ChildProvider{}

func NewChildProvider() ChildProvider { return ChildProvider{} }

type childIdentity struct {
	targetKind       string
	targetID         uuid.UUID
	incidentID       uuid.UUID
	sourceRecordID   uuid.UUID
	sourceFieldKey   string
	indicatorID      uuid.UUID
	resolvedIDs      []uuid.UUID
	expectedVersion  int64
	resolutionStatus string
	resolvedID       *uuid.UUID
	resolvedBy       *uuid.UUID
	resolvedAt       *time.Time
	resolutionMethod *string
	deletedAt        *time.Time
	deletedBy        *uuid.UUID
}

func (p ChildProvider) DescribeTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	identity, before, after, err := parseChildTarget(request.Target)
	if err != nil {
		return rollbackcontract.TargetDescriptor{}, err
	}
	descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: affectedChildRecords(identity, before, after)}
	if request.Target.IncidentID != identity.incidentID {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	if err := validateAffectedRecordsTx(ctx, tx, identity.incidentID, identity, descriptor.AffectedRecordIDs); err != nil {
		return descriptor, err
	}
	current, err := loadChildValueTx(ctx, tx, identity.targetKind, identity.targetID)
	if err != nil {
		return descriptor, err
	}
	currentIdentity, err := parseChildValue(identity.targetKind, current)
	if err != nil {
		return descriptor, err
	}
	if currentIdentity.deletedAt != nil || !sameCurrentChildState(currentIdentity, after) {
		return descriptor, rollbackcontract.ErrStaleTarget
	}
	return descriptor, nil
}

func (p ChildProvider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	descriptor, err := p.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: request.Target})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	identity, beforeState, afterState, err := parseChildTarget(request.Target)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	currentValue, err := loadChildValueTx(ctx, tx, identity.targetKind, identity.targetID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	switch request.Target.OperationKind {
	case "create":
		err = tombstoneChildTx(ctx, tx, identity, request.ActorUserID, request.Now)
	case "resolve":
		err = restoreObservationResolutionTx(ctx, tx, afterState, beforeState)
	default:
		err = rollbackcontract.ErrTargetNotReversible
	}
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	afterValue, err := loadChildValueTx(ctx, tx, identity.targetKind, identity.targetID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return rollbackcontract.ApplyInverseResult{
		AffectedRecordIDs: descriptor.AffectedRecordIDs,
		BeforeValue:       currentValue,
		AfterValue:        afterValue,
		ChangedFieldKeys:  childChangedFieldKeys(identity, beforeState, afterState),
	}, nil
}

func parseChildTarget(target rollbackcontract.NonRowTarget) (childIdentity, childIdentity, childIdentity, error) {
	if target.TargetKind != "indicator_observation" && target.TargetKind != "indicator_state_interval" {
		return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	targetID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotFound
	}
	switch target.OperationKind {
	case "create":
		if target.AfterValue == nil {
			return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		after, err := parseChildValue(target.TargetKind, target.AfterValue)
		if err != nil || after.targetID != targetID {
			return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		return after, childIdentity{}, after, nil
	case "resolve":
		if target.TargetKind != "indicator_observation" || target.BeforeValue == nil || target.AfterValue == nil {
			return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		before, err := parseChildValue(target.TargetKind, target.BeforeValue)
		if err != nil {
			return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		after, err := parseChildValue(target.TargetKind, target.AfterValue)
		if err != nil || before.targetID != targetID || after.targetID != targetID || !sameChildIdentity(before, after) {
			return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		return after, before, after, nil
	default:
		return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
}

func parseChildValue(targetKind string, value map[string]any) (childIdentity, error) {
	identity := childIdentity{targetKind: targetKind}
	var err error
	if identity.incidentID, err = requiredChildUUID(value, "incident_id"); err != nil {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	if identity.expectedVersion, err = requiredChildInt64(value, "row_version"); err != nil || identity.expectedVersion < 1 {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	identity.deletedAt, _, err = optionalChildTime(value, "deleted_at")
	if err != nil {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	identity.deletedBy, _, err = optionalChildUUID(value, "deleted_by_user_id")
	if err != nil || (identity.deletedAt == nil) != (identity.deletedBy == nil) {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	switch targetKind {
	case "indicator_observation":
		if identity.targetID, err = requiredChildUUID(value, "indicator_observation_id"); err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if identity.sourceRecordID, err = requiredChildUUID(value, "source_record_id"); err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		identity.sourceFieldKey = requiredChildText(value, "source_field_key")
		identity.resolutionStatus = requiredChildText(value, "resolution_status")
		if identity.sourceFieldKey == "" || !validObservationStatus(identity.resolutionStatus) {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		for _, key := range []string{"origin_kind", "origin_locator", "observed_text", "created_by_user_id", "created_at"} {
			if requiredChildText(value, key) == "" {
				return childIdentity{}, rollbackcontract.ErrTargetNotReversible
			}
		}
		identity.resolvedID, _, err = optionalChildUUID(value, "resolved_indicator_record_id")
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		identity.resolvedBy, _, err = optionalChildUUID(value, "resolved_by_user_id")
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		identity.resolvedAt, _, err = optionalChildTime(value, "resolved_at")
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		identity.resolutionMethod, _, err = optionalChildText(value, "resolution_method")
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if identity.resolutionStatus == "resolved" {
			if identity.resolvedID == nil || identity.resolvedBy == nil || identity.resolvedAt == nil {
				return childIdentity{}, rollbackcontract.ErrTargetNotReversible
			}
		} else if identity.resolvedID != nil || identity.resolvedBy != nil || identity.resolvedAt != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if identity.resolvedID != nil {
			identity.resolvedIDs = []uuid.UUID{*identity.resolvedID}
		}
	case "indicator_state_interval":
		if identity.targetID, err = requiredChildUUID(value, "indicator_state_interval_id"); err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if identity.indicatorID, err = requiredChildUUID(value, "indicator_record_id"); err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if requiredChildText(value, "lifecycle_state") == "" || requiredChildText(value, "valid_from") == "" || requiredChildText(value, "created_by_user_id") == "" || requiredChildText(value, "created_at") == "" {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		validFrom, err := time.Parse(time.RFC3339Nano, requiredChildText(value, "valid_from"))
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		validTo, _, err := optionalChildTime(value, "valid_to")
		if err != nil || (validTo != nil && validTo.Before(validFrom)) {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if confidence, present, valid := optionalChildInt64(value, "confidence"); !valid || (present && (confidence < 0 || confidence > 100)) {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		if raw, present := value["support_refs"]; present && raw != nil {
			if _, valid := raw.([]any); !valid {
				if _, validStrings := raw.([]string); !validStrings {
					return childIdentity{}, rollbackcontract.ErrTargetNotReversible
				}
			}
		}
		identity.resolvedIDs = []uuid.UUID{identity.indicatorID}
	default:
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	return identity, nil
}

func affectedChildRecords(identity childIdentity, before childIdentity, after childIdentity) []uuid.UUID {
	set := map[uuid.UUID]struct{}{}
	if identity.sourceRecordID != uuid.Nil {
		set[identity.sourceRecordID] = struct{}{}
	}
	if identity.indicatorID != uuid.Nil {
		set[identity.indicatorID] = struct{}{}
	}
	for _, candidate := range append(append([]uuid.UUID(nil), before.resolvedIDs...), after.resolvedIDs...) {
		if candidate != uuid.Nil {
			set[candidate] = struct{}{}
		}
	}
	result := make([]uuid.UUID, 0, len(set))
	for recordID := range set {
		result = append(result, recordID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result
}

func validateAffectedRecordsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, identity childIdentity, recordIDs []uuid.UUID) error {
	rows, err := tx.Query(ctx, `SELECT record_id, record_type FROM records WHERE incident_id = $1 AND record_id = ANY($2::uuid[])`, incidentID, recordIDs)
	if err != nil {
		return err
	}
	defer rows.Close()
	found := map[uuid.UUID]string{}
	for rows.Next() {
		var recordID uuid.UUID
		var recordType string
		if err := rows.Scan(&recordID, &recordType); err != nil {
			return err
		}
		found[recordID] = recordType
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(found) != len(recordIDs) {
		return rollbackcontract.ErrTargetNotFound
	}
	for _, recordID := range recordIDs {
		if identity.sourceRecordID != uuid.Nil && recordID == identity.sourceRecordID {
			continue
		}
		if found[recordID] != "indicator" {
			return rollbackcontract.ErrTargetNotFound
		}
	}
	return nil
}

func loadChildValueTx(ctx context.Context, tx pgx.Tx, targetKind string, targetID uuid.UUID) (map[string]any, error) {
	var raw []byte
	var err error
	switch targetKind {
	case "indicator_observation":
		err = tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'indicator_observation_id', indicator_observation_id::text,
    'incident_id', incident_id::text,
    'source_record_id', source_record_id::text,
    'source_field_key', source_field_key,
    'origin_kind', origin_kind,
    'origin_locator', origin_locator,
    'observed_text', observed_text,
    'parsed_indicator_type', parsed_indicator_type,
    'normalized_candidate', normalized_candidate,
    'resolution_status', resolution_status,
    'resolved_indicator_record_id', resolved_indicator_record_id::text,
    'row_version', row_version,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'resolved_by_user_id', resolved_by_user_id::text,
    'resolved_at', resolved_at,
    'resolution_method', resolution_method,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM indicator_observations
 WHERE indicator_observation_id = $1
`, targetID).Scan(&raw)
	case "indicator_state_interval":
		err = tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'indicator_state_interval_id', indicator_state_interval_id::text,
    'incident_id', incident_id::text,
    'indicator_record_id', indicator_record_id::text,
    'lifecycle_state', lifecycle_state,
    'valid_from', valid_from,
    'valid_to', valid_to,
    'confidence', confidence,
    'rationale', rationale,
    'support_refs', support_refs,
    'assessor', assessor,
    'assessed_at', assessed_at,
    'row_version', row_version,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM indicator_state_intervals
 WHERE indicator_state_interval_id = $1
`, targetID).Scan(&raw)
	default:
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rollbackcontract.ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func tombstoneChildTx(ctx context.Context, tx pgx.Tx, identity childIdentity, actorUserID uuid.UUID, now time.Time) error {
	var rowsAffected int64
	switch identity.targetKind {
	case "indicator_observation":
		tag, err := tx.Exec(ctx, `
UPDATE indicator_observations
   SET deleted_at = $3,
       deleted_by_user_id = $4,
       row_version = row_version + 1
 WHERE indicator_observation_id = $1
   AND row_version = $2
   AND deleted_at IS NULL
`, identity.targetID, identity.expectedVersion, now.UTC(), actorUserID)
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	case "indicator_state_interval":
		tag, err := tx.Exec(ctx, `
UPDATE indicator_state_intervals
   SET deleted_at = $3,
       deleted_by_user_id = $4,
       row_version = row_version + 1
 WHERE indicator_state_interval_id = $1
   AND row_version = $2
   AND deleted_at IS NULL
`, identity.targetID, identity.expectedVersion, now.UTC(), actorUserID)
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	default:
		return rollbackcontract.ErrTargetNotReversible
	}
	if rowsAffected != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func restoreObservationResolutionTx(ctx context.Context, tx pgx.Tx, current childIdentity, retained childIdentity) error {
	tag, err := tx.Exec(ctx, `
UPDATE indicator_observations
   SET resolution_status = $3,
       resolved_indicator_record_id = $4,
       resolved_by_user_id = $5,
       resolved_at = $6,
       resolution_method = $7,
       deleted_at = NULL,
       deleted_by_user_id = NULL,
       row_version = row_version + 1
 WHERE indicator_observation_id = $1
   AND row_version = $2
   AND deleted_at IS NULL
`, current.targetID, current.expectedVersion, retained.resolutionStatus, retained.resolvedID, retained.resolvedBy, retained.resolvedAt, retained.resolutionMethod)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func sameChildIdentity(left childIdentity, right childIdentity) bool {
	return left.targetKind == right.targetKind && left.targetID == right.targetID && left.incidentID == right.incidentID && left.sourceRecordID == right.sourceRecordID && left.sourceFieldKey == right.sourceFieldKey && left.indicatorID == right.indicatorID
}

func sameCurrentChildState(current childIdentity, expected childIdentity) bool {
	if !sameChildIdentity(current, expected) || current.expectedVersion != expected.expectedVersion || current.resolutionStatus != expected.resolutionStatus {
		return false
	}
	return equalUUIDPointers(current.resolvedID, expected.resolvedID) && equalUUIDPointers(current.resolvedBy, expected.resolvedBy) && equalTimePointers(current.resolvedAt, expected.resolvedAt) && equalStringPointers(current.resolutionMethod, expected.resolutionMethod)
}

func childChangedFieldKeys(identity childIdentity, before childIdentity, after childIdentity) map[uuid.UUID][]string {
	changed := map[uuid.UUID][]string{}
	add := func(recordID uuid.UUID, keys ...string) {
		changed[recordID] = append(changed[recordID], keys...)
	}
	if identity.sourceRecordID != uuid.Nil {
		add(identity.sourceRecordID, identity.sourceFieldKey)
	}
	if identity.indicatorID != uuid.Nil {
		add(identity.indicatorID, "indicator.lifecycle_summary")
	}
	for _, indicatorID := range append(append([]uuid.UUID(nil), before.resolvedIDs...), after.resolvedIDs...) {
		add(indicatorID, "indicator.first_observed_at", "indicator.last_observed_at", "indicator.observation_count")
	}
	for recordID := range changed {
		sort.Strings(changed[recordID])
		changed[recordID] = compactStrings(changed[recordID])
	}
	return changed
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func validObservationStatus(value string) bool {
	switch value {
	case "unresolved", "resolved", "dismissed":
		return true
	default:
		return false
	}
}

func requiredChildUUID(value map[string]any, key string) (uuid.UUID, error) {
	raw, valid := value[key].(string)
	if !valid || strings.TrimSpace(raw) == "" {
		return uuid.Nil, errors.New("missing uuid")
	}
	return uuid.Parse(raw)
}

func optionalChildUUID(value map[string]any, key string) (*uuid.UUID, bool, error) {
	raw, present := value[key]
	if !present || raw == nil || raw == "" {
		return nil, present, nil
	}
	text, valid := raw.(string)
	if !valid {
		return nil, true, errors.New("invalid uuid")
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	return &parsed, true, nil
}

func requiredChildInt64(value map[string]any, key string) (int64, error) {
	parsed, present, valid := optionalChildInt64(value, key)
	if !present || !valid {
		return 0, errors.New("invalid integer")
	}
	return parsed, nil
}

func optionalChildInt64(value map[string]any, key string) (int64, bool, bool) {
	raw, present := value[key]
	if !present || raw == nil {
		return 0, present, true
	}
	switch typed := raw.(type) {
	case int:
		return int64(typed), true, true
	case int64:
		return typed, true, true
	case float64:
		return int64(typed), true, typed == float64(int64(typed))
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, true, err == nil
	default:
		return 0, true, false
	}
}

func optionalChildTime(value map[string]any, key string) (*time.Time, bool, error) {
	raw, present := value[key]
	if !present || raw == nil || raw == "" {
		return nil, present, nil
	}
	text, valid := raw.(string)
	if !valid {
		return nil, true, errors.New("invalid time")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return nil, true, err
	}
	utc := parsed.UTC()
	return &utc, true, nil
}

func optionalChildText(value map[string]any, key string) (*string, bool, error) {
	raw, present := value[key]
	if !present || raw == nil || raw == "" {
		return nil, present, nil
	}
	text, valid := raw.(string)
	if !valid {
		return nil, true, errors.New("invalid text")
	}
	return &text, true, nil
}

func requiredChildText(value map[string]any, key string) string {
	text, _ := value[key].(string)
	return strings.TrimSpace(text)
}

func equalUUIDPointers(left *uuid.UUID, right *uuid.UUID) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func equalTimePointers(left *time.Time, right *time.Time) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && left.Equal(*right))
}

func equalStringPointers(left *string, right *string) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}
