package links

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) UpsertFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $7, $4, 'manual', NULL, $5, $5, $6, $6)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type, field_key)
WHERE deleted_at IS NULL AND field_key IS NOT NULL
DO NOTHING
`, incidentID, src, dst, fieldKey, actorID, now, FieldLinkType(fieldKey))
	if err != nil {
		return false, fmt.Errorf("upsert reference link: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) SyncTaskDecisionReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, taskID uuid.UUID, decisionID *uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	changed := false
	args := []any{incidentID, taskID, actorID, now}
	keepPredicate := ""
	if decisionID != nil {
		args = append(args, *decisionID)
		keepPredicate = "AND dst_record_id <> $5"
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $4,
       deleted_by_user_id = $3
 WHERE incident_id = $1
   AND src_record_id = $2
   AND link_type = 'references_record'
   AND field_key = 'task.decision_record_id'
   AND deleted_at IS NULL
   `+keepPredicate, args...)
	if err != nil {
		return false, fmt.Errorf("sync task decision link: %w", err)
	}
	changed = changed || tag.RowsAffected() > 0
	if decisionID == nil {
		return changed, nil
	}
	inserted, err := s.UpsertFieldReferenceTx(ctx, tx, incidentID, taskID, *decisionID, "task.decision_record_id", actorID, now)
	if err != nil {
		return false, err
	}
	return changed || inserted, nil
}

func (s *Store) InsertLinkedNoteReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, actorID uuid.UUID, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_artifact', NULL, 'manual', NULL, $4, $4, $5, $5)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type)
WHERE deleted_at IS NULL AND field_key IS NULL
DO NOTHING
`, incidentID, src, dst, actorID, now)
	if err != nil {
		return fmt.Errorf("insert linked note reference: %w", err)
	}
	return nil
}

func (s *Store) TombstoneFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, expectedTargetType string, actorID uuid.UUID, now time.Time) (bool, error) {
	targetTypePredicate := ""
	args := []any{incidentID, src, dst, fieldKey, actorID, now, FieldLinkType(fieldKey)}
	if expectedTargetType != "" {
		targetTypePredicate = "AND dst.record_type = $8"
		args = append(args, expectedTargetType)
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $6,
       deleted_by_user_id = $5
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $7
   AND field_key = $4
   AND deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM records dst
        WHERE dst.incident_id = record_links.incident_id
          AND dst.record_id = record_links.dst_record_id
          AND dst.deleted_at IS NULL
          `+targetTypePredicate+`
   )
`, args...)
	if err != nil {
		return false, fmt.Errorf("remove reference link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, ErrFieldReferenceNotFound
	}
	return true, nil
}

func (s *Store) UpsertRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, text string, normalized string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text,
    created_by_user_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (handoff_record_id, normalized_risk_ref_text)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, recordID, text, normalized, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert risk ref: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) TombstoneRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, riskRefID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE handoff_risk_refs
   SET deleted_at = $5,
       deleted_by_user_id = $4
 WHERE incident_id = $1
   AND handoff_record_id = $2
   AND risk_ref_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, riskRefID, actorID, now)
	if err != nil {
		return false, fmt.Errorf("remove risk ref: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, ErrRiskRefNotFound
	}
	return true, nil
}

func (s *Store) UpsertTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, actorID uuid.UUID, now time.Time) (bool, error) {
	if tagName == "" || normalizedTagName == "" {
		return false, ErrInvalidTag
	}
	tag, err := tx.Exec(ctx, `
INSERT INTO record_tags (
    incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (incident_id, record_id, normalized_tag_name)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, recordID, tagName, normalizedTagName, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert record tag: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) TombstoneTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = $5,
       deleted_by_user_id = $4,
       updated_at = $5
 WHERE incident_id = $1
   AND record_id = $2
   AND record_tag_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, tagID, actorID, now)
	if err != nil {
		return false, fmt.Errorf("remove record tag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, ErrTagNotFound
	}
	return true, nil
}

func ParseRecordTagItemRef(itemRef string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil || recordID.String() != parts[1] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	tagID, err := uuid.Parse(parts[2])
	if err != nil || tagID.String() != parts[2] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	return recordID, tagID, nil
}

func FieldLinkType(fieldKey string) string {
	switch fieldKey {
	case "decision.support_refs", "finding.supporting_refs":
		return "supported_by"
	default:
		return "references_record"
	}
}
