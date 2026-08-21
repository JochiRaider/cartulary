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

func (Provider) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) TombstoneRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
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

func (Provider) RestoreRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, value map[string]any, actorUserID uuid.UUID, now time.Time) error {
	plan, err := valuecodec.DecodeRecordLinkRestorePlan(value, actorUserID)
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
       decided_at = COALESCE($10, decided_at),
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_link_id = $1
   AND incident_id = $2
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, plan.FieldKeyValue(), plan.Provenance, plan.Confidence, plan.OwnerUserID, plan.DecidedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, COALESCE($10, $11), $11)
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, plan.FieldKeyValue(), plan.Provenance, plan.Confidence, plan.OwnerUserID, plan.DecidedAt, now.UTC())
	return err
}

func (Provider) LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) RestoreRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, value map[string]any, now time.Time) error {
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
       updated_at = $5,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_tag_id = $1
   AND incident_id = $6
`, recordTagID, identity.RecordID, plan.TagName, plan.NormalizedTagName, now.UTC(), identity.IncidentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrTargetNotFound
	}
	return nil
}

func (Provider) TombstoneRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
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
	value := target.BeforeValue
	if target.OperationKind == "create" || value == nil {
		value = target.AfterValue
	}
	switch target.TargetKind {
	case "record_link":
		if target.OperationKind != "create" && target.OperationKind != "delete" {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		parsed, err := valuecodec.DecodeRecordLinkMutationValue(value)
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
		}
		if parsed.IncidentID != target.IncidentID || parsed.RecordLinkID.String() != target.TargetID {
			return rollbackcontract.TargetDescriptor{}, ErrTargetNotFound
		}
		descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: canonicalIDs(parsed.SrcRecordID, parsed.DstRecordID)}
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
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM records WHERE incident_id = $1 AND record_id = ANY($2::uuid[])`, target.IncidentID, []uuid.UUID{parsed.SrcRecordID, parsed.DstRecordID}).Scan(&endpointCount); err != nil {
			return descriptor, err
		}
		wantEndpoints := 2
		if parsed.SrcRecordID == parsed.DstRecordID {
			wantEndpoints = 1
		}
		if endpointCount != wantEndpoints {
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
		descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: []uuid.UUID{parsed.RecordID}}
		var recordExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $2)`, target.IncidentID, parsed.RecordID).Scan(&recordExists); err != nil {
			return descriptor, err
		}
		if !recordExists {
			return descriptor, ErrTargetNotFound
		}
		return descriptor, nil
	default:
		return rollbackcontract.TargetDescriptor{}, ErrTargetNotReversible
	}
}

func (p Provider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
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
		before, err = p.LoadRecordLinkValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		switch request.Target.OperationKind {
		case "create":
			err = p.TombstoneRecordLinkTx(ctx, tx, request.Target.IncidentID, targetID, request.ActorUserID, request.Now)
		case "delete":
			err = p.RestoreRecordLinkTx(ctx, tx, request.Target.IncidentID, targetID, request.Target.BeforeValue, request.ActorUserID, request.Now)
		}
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		after, err := p.LoadRecordLinkValueTx(ctx, tx, targetID)
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
		before, err = p.LoadRecordTagValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		switch request.Target.OperationKind {
		case "create":
			err = p.TombstoneRecordTagTx(ctx, tx, targetID, request.ActorUserID, request.Now)
		case "patch", "delete":
			err = p.RestoreRecordTagTx(ctx, tx, targetID, request.Target.BeforeValue, request.Now)
		}
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		after, err := p.LoadRecordTagValueTx(ctx, tx, targetID)
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		changed, err := recordTagChangedFieldKeysTx(ctx, tx, descriptor.AffectedRecordIDs[0])
		if err != nil {
			return rollbackcontract.ApplyInverseResult{}, err
		}
		return rollbackcontract.ApplyInverseResult{AffectedRecordIDs: descriptor.AffectedRecordIDs, BeforeValue: before, AfterValue: after, ChangedFieldKeys: changed}, nil
	default:
		return rollbackcontract.ApplyInverseResult{}, ErrTargetNotReversible
	}
}

func matchesRecordTagTargetID(targetID string, recordID uuid.UUID, recordTagID uuid.UUID) bool {
	if targetID == recordTagID.String() {
		return true
	}
	return targetID == "record_tag:"+recordID.String()+":"+recordTagID.String()
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

func recordTagChangedFieldKeysTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[uuid.UUID][]string, error) {
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
	return map[uuid.UUID][]string{recordID: keys}, nil
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
