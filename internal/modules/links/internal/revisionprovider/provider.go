package revisionprovider

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

var (
	ErrTargetNotFound      = rollbackcontract.ErrTargetNotFound
	ErrStaleTarget         = rollbackcontract.ErrStaleTarget
	ErrTargetNotReversible = rollbackcontract.ErrTargetNotReversible
)

type Provider struct{}

var _ rollbackcontract.NonRowTargetProvider = Provider{}

type RecordTagIdentity = valuecodec.RecordTagIdentity

func NewProvider() Provider {
	return Provider{}
}

func (Provider) ValidateRecordLinkValue(value map[string]any) error {
	_, err := valuecodec.DecodeRecordLinkMutationValue(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	return nil
}

func (Provider) ParseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	parsed, err := valuecodec.DecodeRecordTagMutationValue(value)
	if err != nil {
		return RecordTagIdentity{}, ErrTargetNotReversible
	}
	return RecordTagIdentity{RecordTagID: parsed.RecordTagID, IncidentID: parsed.IncidentID, RecordID: parsed.RecordID}, nil
}

func (Provider) loadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) tombstoneRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $3,
       deleted_by_user_id = $4
 WHERE record_link_id = $1
   AND incident_id = $2
   AND deleted_at IS NULL
`, recordLinkID, incidentID, now.UTC(), actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleTarget
	}
	return nil
}

func (Provider) restoreRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, value map[string]any) error {
	plan, err := valuecodec.DecodeRecordLinkRestorePlan(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	identity := plan.Identity
	if identity.IncidentID != incidentID || identity.RecordLinkID != recordLinkID {
		return ErrTargetNotFound
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET src_record_id = $3,
       dst_record_id = $4,
       link_type = $5,
       field_key = $6,
       provenance = $7,
	       confidence = $8,
	       owner_user_id = $9,
	       created_by_user_id = $10,
	       decided_at = $11,
	       created_at = $12,
	       deleted_at = $13,
	       deleted_by_user_id = $14
 WHERE record_link_id = $1
   AND incident_id = $2
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, plan.FieldKeyValue(), plan.Provenance, plan.Confidence, plan.OwnerUserID, plan.CreatedByUserID, plan.DecidedAt, plan.CreatedAt, plan.DeletedAt, plan.DeletedByUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at,
    deleted_at, deleted_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, plan.FieldKeyValue(), plan.Provenance, plan.Confidence, plan.OwnerUserID, plan.CreatedByUserID, plan.DecidedAt, plan.CreatedAt, plan.DeletedAt, plan.DeletedByUserID)
	return err
}

func (Provider) loadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) restoreRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, value map[string]any) error {
	plan, err := valuecodec.DecodeRecordTagRestorePlan(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	identity := plan.Identity
	if identity.RecordTagID != recordTagID {
		return ErrTargetNotFound
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET record_id = $2,
	       tag_name = $3,
	       normalized_tag_name = $4,
	       created_by_user_id = $5,
	       created_at = $6,
	       updated_at = $7,
	       deleted_at = $8,
	       deleted_by_user_id = $9
 WHERE record_tag_id = $1
	AND incident_id = $10
`, recordTagID, identity.RecordID, plan.TagName, plan.NormalizedTagName, plan.CreatedByUserID, plan.CreatedAt, plan.UpdatedAt, plan.DeletedAt, plan.DeletedByUserID, identity.IncidentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO record_tags (
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at, deleted_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, recordTagID, identity.IncidentID, identity.RecordID, plan.TagName, plan.NormalizedTagName, plan.CreatedByUserID, plan.CreatedAt, plan.UpdatedAt, plan.DeletedAt, plan.DeletedByUserID)
	return err
}

func (Provider) tombstoneRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2
 WHERE record_tag_id = $1
   AND deleted_at IS NULL
`, recordTagID, now.UTC(), actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleTarget
	}
	return nil
}

func (p Provider) DescribeTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	target := request.Target
	if err := validateRetainedTarget(target); err != nil {
		return rollbackcontract.TargetDescriptor{}, err
	}
	value := target.BeforeValue
	if target.OperationKind == "create" || value == nil {
		value = target.AfterValue
	}
	switch target.TargetKind {
	case "record_link":
		if target.OperationKind != "create" && target.OperationKind != "patch" && target.OperationKind != "delete" {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		parsed, err := valuecodec.DecodeRecordLinkMutationValue(value)
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		if parsed.IncidentID != target.IncidentID || parsed.RecordLinkID.String() != target.TargetID {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotFound
		}
		affected := []uuid.UUID{parsed.SrcRecordID, parsed.DstRecordID}
		if target.AfterValue != nil {
			after, decodeErr := valuecodec.DecodeRecordLinkMutationValue(target.AfterValue)
			if decodeErr != nil {
				return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
			}
			affected = append(affected, after.SrcRecordID, after.DstRecordID)
		}
		descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: canonicalIDs(affected...)}
		if parsed.LinkType == "attached_evidence" && target.ChangeSetID != uuid.Nil {
			for _, sibling := range request.SiblingTargets {
				if sibling.DispatchClass == rollbackcontract.DispatchRow {
					descriptor.RequiresWholeChangeSetWith = append(
						descriptor.RequiresWholeChangeSetWith,
						sibling.TargetReference,
					)
				}
			}
		}
		var endpointCount int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM records WHERE incident_id = $1 AND record_id = ANY($2::uuid[])`, target.IncidentID, descriptor.AffectedRecordIDs).Scan(&endpointCount); err != nil {
			return descriptor, err
		}
		if endpointCount != len(descriptor.AffectedRecordIDs) {
			return descriptor, ErrTargetNotFound
		}
		return descriptor, nil
	case "record_tag":
		if target.OperationKind != "create" && target.OperationKind != "patch" && target.OperationKind != "delete" {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		parsed, err := valuecodec.DecodeRecordTagMutationValue(value)
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		if parsed.IncidentID != target.IncidentID || !matchesRecordTagTargetID(target.TargetID, parsed.RecordID, parsed.RecordTagID) {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotFound
		}
		affected := []uuid.UUID{parsed.RecordID}
		if target.AfterValue != nil {
			after, decodeErr := valuecodec.DecodeRecordTagMutationValue(target.AfterValue)
			if decodeErr != nil {
				return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
			}
			affected = append(affected, after.RecordID)
		}
		descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: canonicalIDs(affected...)}
		var recordCount int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM records WHERE incident_id = $1 AND record_id = ANY($2::uuid[])`, target.IncidentID, descriptor.AffectedRecordIDs).Scan(&recordCount); err != nil {
			return descriptor, err
		}
		if recordCount != len(descriptor.AffectedRecordIDs) {
			return descriptor, ErrTargetNotFound
		}
		return descriptor, nil
	default:
		return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
	}
}

func (p Provider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	if err := validateRetainedTarget(request.Target); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	descriptor, err := p.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: request.Target})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	var before map[string]any
	switch request.Target.TargetKind {
	case "record_link":
		targetID, err := uuid.Parse(request.Target.TargetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, ErrTargetNotFound
		}
		before, err = p.loadRecordLinkValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		switch request.Target.OperationKind {
		case "create":
			err = p.tombstoneRecordLinkTx(ctx, tx, request.Target.IncidentID, targetID, request.ActorUserID, request.Now)
		case "patch", "delete":
			err = p.restoreRecordLinkTx(ctx, tx, request.Target.IncidentID, targetID, request.Target.BeforeValue)
		}
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		after, err := p.loadRecordLinkValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		changed, err := recordLinkChangedFieldKeysTx(ctx, tx, request.Target, descriptor.AffectedRecordIDs)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		return rollbackcontract.ApplyInverseResult{AffectedRecordIDs: descriptor.AffectedRecordIDs, BeforeValue: before, AfterValue: after, ChangedFieldKeys: changed}, nil
	case "record_tag":
		value := request.Target.BeforeValue
		if request.Target.OperationKind == "create" || value == nil {
			value = request.Target.AfterValue
		}
		parsed, err := valuecodec.DecodeRecordTagMutationValue(value)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, ErrTargetNotReversible
		}
		targetID := parsed.RecordTagID
		before, err = p.loadRecordTagValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		switch request.Target.OperationKind {
		case "create":
			err = p.tombstoneRecordTagTx(ctx, tx, targetID, request.ActorUserID, request.Now)
		case "patch", "delete":
			err = p.restoreRecordTagTx(ctx, tx, targetID, request.Target.BeforeValue)
		}
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		after, err := p.loadRecordTagValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		changed, err := recordTagChangedFieldKeysTx(ctx, tx, descriptor.AffectedRecordIDs)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		return rollbackcontract.ApplyInverseResult{AffectedRecordIDs: descriptor.AffectedRecordIDs, BeforeValue: before, AfterValue: after, ChangedFieldKeys: changed}, nil
	default:
		return rollbackcontract.ApplyInverseResult{}, ErrTargetNotReversible
	}
}

func matchesRecordTagTargetID(targetID string, recordID uuid.UUID, recordTagID uuid.UUID) bool {
	return targetID == "record_tag:"+recordID.String()+":"+recordTagID.String()
}

func validateRetainedTarget(target rollbackcontract.NonRowTarget) error {
	if err := valuecodec.ValidateHistoryMutation(
		target.TargetKind,
		target.TargetID,
		target.OperationKind,
		target.BeforeValue,
		target.AfterValue,
	); err != nil {
		return ErrTargetNotReversible
	}
	return nil
}

func recordLinkChangedFieldKeysTx(ctx context.Context, tx pgx.Tx, target rollbackcontract.NonRowTarget, affected []uuid.UUID) (map[uuid.UUID][]string, error) {
	value := target.BeforeValue
	if target.OperationKind == "create" || value == nil {
		value = target.AfterValue
	}
	parsed, err := valuecodec.DecodeRecordLinkMutationValue(value)
	if err != nil {
		return nil, ErrTargetNotReversible
	}
	fieldKey, _ := value["field_key"].(string)
	changed := make(map[uuid.UUID][]string, len(affected))
	for _, recordID := range affected {
		var recordType string
		if err := tx.QueryRow(ctx, `SELECT record_type FROM records WHERE incident_id = $1 AND record_id = $2`, target.IncidentID, recordID).Scan(&recordType); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrTargetNotFound
			}
			return nil, err
		}
		keys := []string{}
		switch {
		case parsed.LinkType == "attached_evidence" && recordID == parsed.SrcRecordID:
			switch recordType {
			case "timeline_event":
				keys = append(keys, "timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence")
			case "host":
				keys = append(keys, "host.evidence_count")
			case "identity":
				keys = append(keys, "identity.evidence_count")
			}
		case parsed.LinkType == "attached_evidence" && recordID == parsed.DstRecordID && recordType == "evidence":
			keys = append(keys, "evidence.linked_record_count")
		case parsed.LinkType == "supersedes" && recordID == parsed.SrcRecordID && recordType == "decision":
			keys = append(keys, "decision.supersedes_record_id")
		case parsed.LinkType == "supersedes" && recordID == parsed.DstRecordID && recordType == "decision":
			keys = append(keys, "decision.is_superseded")
		case recordID == parsed.SrcRecordID && strings.TrimSpace(fieldKey) != "":
			keys = append(keys, fieldKey)
		}
		sort.Strings(keys)
		changed[recordID] = keys
	}
	return changed, nil
}

func recordTagChangedFieldKeysTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	changed := make(map[uuid.UUID][]string, len(recordIDs))
	for _, recordID := range recordIDs {
		var recordType string
		if err := tx.QueryRow(ctx, `SELECT record_type FROM records WHERE record_id = $1`, recordID).Scan(&recordType); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrTargetNotFound
			}
			return nil, err
		}
		keys := []string{}
		switch recordType {
		case "timeline_event":
			keys = append(keys, "timeline.tags")
		case "artifact":
			keys = append(keys, "note.tags")
		}
		changed[recordID] = keys
	}
	return changed, nil
}

func canonicalIDs(values ...uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			set[value] = struct{}{}
		}
	}
	result := make([]uuid.UUID, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result
}
