package rollback

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/intervalhistory"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/observationhistory"
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
	if !sameCompleteChildState(identity.targetKind, current, request.Target.AfterValue) {
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
	}, nil
}

func parseChildTarget(target rollbackcontract.NonRowTarget) (childIdentity, childIdentity, childIdentity, error) {
	if target.TargetKind == "indicator_observation" && ((target.OperationKind != "create" && target.OperationKind != "resolve") || observationhistory.ValidateMutation(target.TargetID, target.OperationKind, target.BeforeValue, target.AfterValue) != nil) {
		return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	if target.TargetKind == "indicator_state_interval" && (target.OperationKind != "create" || intervalhistory.ValidateMutation(target.TargetID, target.OperationKind, target.BeforeValue, target.AfterValue) != nil) {
		return childIdentity{}, childIdentity{}, childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}

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
	if targetKind == "indicator_state_interval" {
		v, err := intervalhistory.Decode(value)
		if err != nil {
			return childIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
		return childIdentity{targetKind: targetKind, targetID: v.IntervalID, incidentID: v.IncidentID, indicatorID: v.IndicatorRecordID,
			expectedVersion: v.RowVersion, deletedAt: v.DeletedAt, deletedBy: v.DeletedByUserID, resolvedIDs: []uuid.UUID{v.IndicatorRecordID}}, nil
	}

	if targetKind != "indicator_observation" {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	v, err := observationhistory.Decode(value)
	if err != nil {
		return childIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	identity := childIdentity{targetKind: targetKind, targetID: v.ObservationID, incidentID: v.IncidentID, sourceRecordID: v.SourceRecordID, sourceFieldKey: v.SourceFieldKey,
		expectedVersion: v.RowVersion, resolutionStatus: v.ResolutionStatus, resolvedID: v.ResolvedIndicatorRecordID, resolvedBy: v.ResolvedByUserID, resolvedAt: v.ResolvedAt, resolutionMethod: v.ResolutionMethod, deletedAt: v.DeletedAt, deletedBy: v.DeletedByUserID}
	if v.ResolvedIndicatorRecordID != nil {
		identity.resolvedIDs = []uuid.UUID{*v.ResolvedIndicatorRecordID}
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
	var value map[string]any
	var err error
	switch targetKind {
	case "indicator_observation":
		var v observationhistory.Value
		v, err = observationhistory.Scan(tx.QueryRow(ctx, `SELECT indicator_observation_id, incident_id, source_record_id, source_field_key, origin_kind, origin_locator, observed_text, parsed_indicator_type, normalized_candidate, resolution_status, resolved_indicator_record_id, row_version, created_by_user_id, created_at, resolved_by_user_id, resolved_at, resolution_method, deleted_at, deleted_by_user_id FROM indicator_observations WHERE indicator_observation_id=$1`, targetID))
		if err == nil {
			value = v.Map()
		}
	case "indicator_state_interval":
		var v intervalhistory.Value
		v, err = intervalhistory.Scan(tx.QueryRow(ctx, `SELECT indicator_state_interval_id,incident_id,indicator_record_id,lifecycle_state,
valid_from,valid_to,confidence,rationale,support_refs,assessor,assessed_at,row_version,
created_by_user_id,created_at,deleted_at,deleted_by_user_id
FROM indicator_state_intervals WHERE indicator_state_interval_id=$1`, targetID))
		if err == nil {
			value = v.Map()
		}
	default:
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rollbackcontract.ErrTargetNotFound
	}
	if errors.Is(err, intervalhistory.ErrInvalidValue) || errors.Is(err, observationhistory.ErrInvalidValue) {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	return value, err
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

func sameCompleteChildState(kind string, current, expected map[string]any) bool {
	switch kind {
	case "indicator_observation":
		left, leftErr := observationhistory.Decode(current)
		right, rightErr := observationhistory.Decode(expected)
		return leftErr == nil && rightErr == nil && reflect.DeepEqual(left, right)
	case "indicator_state_interval":
		left, leftErr := intervalhistory.Decode(current)
		right, rightErr := intervalhistory.Decode(expected)
		return leftErr == nil && rightErr == nil && reflect.DeepEqual(left, right)
	default:
		return false
	}
}

func equalStringPointers(left *string, right *string) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}
