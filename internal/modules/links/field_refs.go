package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func listActiveFieldReferenceStatesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, fieldKey string, linkType LinkType) ([]recordLinkState, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
 ORDER BY record_link_id
 FOR UPDATE
`, incidentID, srcRecordID, fieldKey, linkType.String())
	if err != nil {
		return nil, fmt.Errorf("list active field-reference states: %w", err)
	}
	defer rows.Close()
	states := make([]recordLinkState, 0)
	for rows.Next() {
		state, err := scanRecordLinkState(rows)
		if err != nil {
			return nil, fmt.Errorf("scan active field-reference state: %w", err)
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active field-reference states: %w", err)
	}
	return states, nil
}

func upsertFieldReferenceStateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, fieldKey string, linkType LinkType, actorUserID uuid.UUID, now time.Time) (recordLinkState, bool, error) {
	if err := validateRecordLinkCommand(linkType, LinkProvenanceManual, nil, srcRecordID, dstRecordID); err != nil {
		return recordLinkState{}, false, err
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, incidentID, srcRecordID, dstRecordID); err != nil {
		return recordLinkState{}, false, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at,
    created_at
) VALUES ($1, $2, $3, $4, $5, 'manual', NULL, $6, $6, $7, $7)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type, field_key)
WHERE deleted_at IS NULL AND field_key IS NOT NULL
DO NOTHING
RETURNING
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
`, incidentID, srcRecordID, dstRecordID, linkType.String(), fieldKey, actorUserID, now.UTC())
	state, err := scanRecordLinkState(row)
	if err == nil {
		return state, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return recordLinkState{}, false, fmt.Errorf("upsert field-reference link: %w", err)
	}
	state, err = getActiveFieldReferenceStateTx(ctx, tx, incidentID, srcRecordID, dstRecordID, fieldKey, linkType)
	if err != nil {
		return recordLinkState{}, false, err
	}
	return state, false, nil
}

func getActiveFieldReferenceStateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, fieldKey string, linkType LinkType) (recordLinkState, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND field_key = $5
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, srcRecordID, dstRecordID, linkType.String(), fieldKey)
	state, err := scanRecordLinkState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordLinkState{}, errFieldReferenceNotFound
	}
	if err != nil {
		return recordLinkState{}, fmt.Errorf("get active field-reference state: %w", err)
	}
	return state, nil
}

func tombstoneFieldReferenceStateTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, expectedTargetType string, actorUserID uuid.UUID, now time.Time) (recordLinkState, error) {
	targetTypePredicate := ""
	args := []any{recordLinkID, actorUserID, now.UTC()}
	if expectedTargetType != "" {
		targetTypePredicate = "AND dst.record_type = $4"
		args = append(args, expectedTargetType)
	}
	row := tx.QueryRow(ctx, `
UPDATE record_links
   SET deleted_at = $3,
       deleted_by_user_id = $2
 WHERE record_link_id = $1
   AND deleted_at IS NULL
   AND field_key IS NOT NULL
   AND EXISTS (
       SELECT 1
         FROM records dst
        WHERE dst.incident_id = record_links.incident_id
          AND dst.record_id = record_links.dst_record_id
          AND dst.deleted_at IS NULL
          `+targetTypePredicate+`
   )
RETURNING
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
`, args...)
	state, err := scanRecordLinkState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordLinkState{}, errFieldReferenceNotFound
	}
	if err != nil {
		return recordLinkState{}, fmt.Errorf("tombstone field-reference state: %w", err)
	}
	return state, nil
}
