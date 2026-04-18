package entities

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrInvalidMentionResolution = errors.New("entities: invalid mention resolution")

func (s *Store) ResolveOrCreateFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) (MentionResolutionResult, error) {
	mention, err := loadMentionForResolutionTx(ctx, tx, mentionID)
	if err != nil {
		return MentionResolutionResult{}, err
	}
	if mention.SourceRecordID != sourceRecordID || mention.SourceFieldKey != fieldKey || mention.ResolutionStatus != "unresolved" {
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	var (
		result           MentionResolutionResult
		resolutionMethod = "created_from_mention"
	)
	switch mention.EntityType {
	case "host":
		if resolvedRecordID != nil {
			record, err := loadActiveHostByIDTx(ctx, tx, mention.IncidentID, *resolvedRecordID)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			result = MentionResolutionResult{
				EntityType: "host",
				RecordID:   record.RecordID,
			}
			resolutionMethod = "resolve_item"
		} else {
			record, beforeRow, operationKind, _, err := s.upsertHostWithInputTx(ctx, tx, actor, mention.IncidentID, hostInputFromMention(mention), now)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			afterRow := BuildHostRow(record)
			if beforeRow != nil && reflect.DeepEqual(beforeRow, afterRow) {
				operationKind = ""
			}
			result = MentionResolutionResult{
				EntityType:    "host",
				RecordID:      record.RecordID,
				OperationKind: operationKind,
				BeforeRow:     beforeRow,
				AfterRow:      afterRow,
			}
		}
	case "identity":
		if resolvedRecordID != nil {
			record, err := loadActiveIdentityByIDTx(ctx, tx, mention.IncidentID, *resolvedRecordID)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			result = MentionResolutionResult{
				EntityType: "identity",
				RecordID:   record.RecordID,
			}
			resolutionMethod = "resolve_item"
		} else {
			record, beforeRow, operationKind, _, err := s.upsertIdentityWithInputTx(ctx, tx, actor, mention.IncidentID, identityInputFromMention(mention), now)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			afterRow := BuildIdentityRow(record)
			if beforeRow != nil && reflect.DeepEqual(beforeRow, afterRow) {
				operationKind = ""
			}
			result = MentionResolutionResult{
				EntityType:    "identity",
				RecordID:      record.RecordID,
				OperationKind: operationKind,
				BeforeRow:     beforeRow,
				AfterRow:      afterRow,
			}
		}
	default:
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	if _, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET resolution_status = 'resolved',
       row_version = row_version + 1,
       resolved_record_id = $2,
       resolved_by_user_id = $3,
       resolved_at = $4,
       resolution_method = $5
 WHERE entity_mention_id = $1
`, mentionID, result.RecordID, actor.ID, now.UTC(), resolutionMethod); err != nil {
		return MentionResolutionResult{}, fmt.Errorf("resolve mention from item action: %w", err)
	}
	return result, nil
}

func loadMentionForResolutionTx(ctx context.Context, tx pgx.Tx, mentionID uuid.UUID) (mentionRecord, error) {
	var mention mentionRecord
	if err := tx.QueryRow(ctx, `
SELECT
    m.entity_mention_id,
    m.source_record_id,
    t.incident_id,
    m.entity_type,
    m.source_field_key,
    m.raw_text,
    m.resolution_status
  FROM entity_mentions m
  JOIN timeline_events t ON t.record_id = m.source_record_id
 WHERE m.entity_mention_id = $1
 FOR UPDATE
`, mentionID).Scan(
		&mention.EntityMentionID,
		&mention.SourceRecordID,
		&mention.IncidentID,
		&mention.EntityType,
		&mention.SourceFieldKey,
		&mention.RawText,
		&mention.ResolutionStatus,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mentionRecord{}, ErrInvalidMentionResolution
		}
		return mentionRecord{}, fmt.Errorf("load mention for resolution: %w", err)
	}
	return mention, nil
}

func loadActiveHostByIDTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM hosts
 WHERE incident_id = $1
   AND record_id = $2
   AND host_state IN ('stub', 'canonical')
 FOR UPDATE
`, incidentID, recordID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return HostRecord{}, ErrInvalidMentionResolution
		}
		return HostRecord{}, fmt.Errorf("load host by id: %w", err)
	}
	return record, nil
}

func loadActiveIdentityByIDTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (IdentityRecord, error) {
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email::text,
    sam_account_name,
    identity_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM identities
 WHERE incident_id = $1
   AND record_id = $2
   AND identity_state IN ('stub', 'canonical')
 FOR UPDATE
`, incidentID, recordID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return IdentityRecord{}, ErrInvalidMentionResolution
		}
		return IdentityRecord{}, fmt.Errorf("load identity by id: %w", err)
	}
	return record, nil
}
