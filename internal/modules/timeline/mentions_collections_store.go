package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type mentionQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type projectedMentionItem struct {
	MentionID        uuid.UUID
	EntityType       string
	SourceFieldKey   string
	RawText          string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolutionMethod *string
}

func hydrateProjectedCollections(ctx context.Context, querier mentionQueryer, record *projectedRecord) error {
	if record == nil {
		return nil
	}
	rows, err := querier.Query(ctx, `
SELECT entity_mention_id, entity_type, source_field_key, raw_text, resolution_status, row_version, resolved_record_id, resolution_method, ordinal
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status IN ('unresolved', 'resolved')
 ORDER BY ordinal ASC, entity_mention_id ASC
`, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline mention collections: %w", err)
	}

	mentions := make([]projectedMentionItem, 0)
	for rows.Next() {
		var (
			mentionID        uuid.UUID
			entityType       string
			sourceFieldKey   string
			rawText          string
			resolutionStatus string
			rowVersion       int64
			resolvedRecordID pgtype.UUID
			resolutionMethod pgtype.Text
			ordinal          int
		)
		if err := rows.Scan(&mentionID, &entityType, &sourceFieldKey, &rawText, &resolutionStatus, &rowVersion, &resolvedRecordID, &resolutionMethod, &ordinal); err != nil {
			rows.Close()
			return fmt.Errorf("scan timeline mention collection row: %w", err)
		}
		mention := projectedMentionItem{
			MentionID:        mentionID,
			EntityType:       entityType,
			SourceFieldKey:   sourceFieldKey,
			RawText:          rawText,
			ResolutionStatus: resolutionStatus,
			RowVersion:       rowVersion,
		}
		if resolvedRecordID.Valid {
			resolved := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
			mention.ResolvedRecordID = &resolved
		}
		if resolutionMethod.Valid {
			value := resolutionMethod.String
			mention.ResolutionMethod = &value
		}
		mentions = append(mentions, mention)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate timeline mention collection rows: %w", err)
	}
	rows.Close()

	hostRefs := make([]map[string]any, 0)
	identityRefs := make([]map[string]any, 0)
	hasUnresolved := false
	for _, mention := range mentions {
		item := map[string]any{
			"item_ref":            "entity_mention:" + mention.MentionID.String(),
			"entity_type":         mention.EntityType,
			"display_text":        mention.RawText,
			"raw_text":            mention.RawText,
			"mention_row_version": mention.RowVersion,
		}
		if mention.ResolutionStatus == "resolved" && mention.ResolvedRecordID != nil {
			item["item_kind"] = "resolved_ref"
			item["resolved_record_id"] = mention.ResolvedRecordID.String()
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod != "" {
				item["resolution_method"] = *mention.ResolutionMethod
				if *mention.ResolutionMethod == autoResolutionMethod {
					item["auto_resolved"] = true
				}
			}
			if linkType, ok := timelineRelationshipLinkType(mention.SourceFieldKey); ok {
				linkMetadata, err := loadActiveCollectionLinkMetadata(ctx, querier, record.IncidentID, record.RecordID, *mention.ResolvedRecordID, linkType)
				if err != nil {
					return err
				}
				if linkMetadata != nil {
					item["provenance"] = linkMetadata.Provenance
					item["confidence"] = linkMetadata.Confidence
				}
			}
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod == autoResolutionMethod {
				matchedAliasText, err := lookupMatchedAliasText(ctx, querier, *mention.ResolvedRecordID, mention.EntityType, mention.RawText)
				if err != nil {
					return err
				}
				if matchedAliasText != nil {
					item["matched_alias_text"] = *matchedAliasText
				}
			}
		} else {
			item["item_kind"] = "unresolved_mention"
			hasUnresolved = true
		}

		switch mention.EntityType {
		case "host":
			hostRefs = append(hostRefs, item)
		case "identity":
			identityRefs = append(identityRefs, item)
		}
	}
	record.HostRefs = hostRefs
	record.IdentityRefs = identityRefs
	record.HasUnresolvedMentions = hasUnresolved
	tagRows, err := querier.Query(ctx, `
SELECT record_tag_id, tag_name
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline tags: %w", err)
	}

	tags := make([]map[string]any, 0)
	for tagRows.Next() {
		var (
			recordTagID uuid.UUID
			tagName     string
		)
		if err := tagRows.Scan(&recordTagID, &tagName); err != nil {
			tagRows.Close()
			return fmt.Errorf("scan timeline tag row: %w", err)
		}
		tags = append(tags, map[string]any{
			"item_ref":     recordTagMutationTarget(record.RecordID, recordTagID),
			"item_kind":    "tag",
			"display_text": tagName,
			"tag_id":       recordTagID.String(),
		})
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return fmt.Errorf("iterate timeline tags: %w", err)
	}
	tagRows.Close()
	record.Tags = tags

	evidenceRows, err := querier.Query(ctx, `
SELECT
    rl.dst_record_id,
    COALESCE(ev.title, rl.dst_record_id::text) AS title,
    ev.lifecycle_state,
    COALESCE(b.upload_state, ev.upload_state, 'pending') AS upload_state
  FROM record_links rl
  JOIN evidence ev
    ON ev.incident_id = rl.incident_id
   AND ev.record_id = rl.dst_record_id
  LEFT JOIN object_blobs b
    ON b.object_blob_id = ev.object_blob_id
 WHERE rl.incident_id = $1
   AND rl.src_record_id = $2
   AND rl.link_type = 'attached_evidence'
   AND rl.deleted_at IS NULL
 ORDER BY COALESCE(ev.title, rl.dst_record_id::text) ASC, rl.dst_record_id ASC
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline attached evidence: %w", err)
	}
	attachedEvidence := make([]map[string]any, 0)
	availableEvidenceCount := 0
	for evidenceRows.Next() {
		var (
			evidenceRecordID uuid.UUID
			title            string
			lifecycleState   string
			uploadState      string
		)
		if err := evidenceRows.Scan(&evidenceRecordID, &title, &lifecycleState, &uploadState); err != nil {
			evidenceRows.Close()
			return fmt.Errorf("scan timeline attached evidence row: %w", err)
		}
		attachedEvidence = append(attachedEvidence, map[string]any{
			"item_ref":         "record_ref:" + evidenceRecordID.String(),
			"item_kind":        "record_ref",
			"display_text":     title,
			"linked_record_id": evidenceRecordID.String(),
		})
		if uploadState == "available" && (lifecycleState == "available" || lifecycleState == "released") {
			availableEvidenceCount += 1
		}
	}
	if err := evidenceRows.Err(); err != nil {
		evidenceRows.Close()
		return fmt.Errorf("iterate timeline attached evidence rows: %w", err)
	}
	evidenceRows.Close()
	record.AttachedEvidence = attachedEvidence
	record.EvidenceCount = availableEvidenceCount
	record.HasEvidence = availableEvidenceCount > 0
	return nil
}

type mentionInsertOptions struct {
	allowInteractiveAutoResolution bool
	originKind                     string
}

type mentionProjectionRefresh struct {
	Hosts      bool
	Identities bool
}

func (s *Store) rebuildMentionEntityProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, refresh mentionProjectionRefresh) error {
	if refresh.Hosts {
		if err := s.projectionStore.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
			return err
		}
	}
	if refresh.Identities {
		if err := s.projectionStore.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
			return err
		}
	}
	return nil
}

func (r *mentionProjectionRefresh) include(fieldKey string) {
	switch fieldKey {
	case "timeline.host_refs":
		r.Hosts = true
	case "timeline.identity_refs":
		r.Identities = true
	}
}

func (s *Store) applyCreateMentionActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, hostRefs *CollectionActionPayload, identityRefs *CollectionActionPayload, options createRowOptions, now time.Time) (mentionProjectionRefresh, error) {
	var refresh mentionProjectionRefresh
	insertOptions := mentionInsertOptions{
		allowInteractiveAutoResolution: options.allowInteractiveAutoResolution,
	}
	hostLinked, err := insertMentionActionsTx(ctx, tx, s.linkStore, actorUserID, incidentID, recordID, "timeline.host_refs", "host", hostRefs, insertOptions, now)
	if err != nil {
		return mentionProjectionRefresh{}, err
	}
	if hostLinked {
		refresh.Hosts = true
	}
	identityLinked, err := insertMentionActionsTx(ctx, tx, s.linkStore, actorUserID, incidentID, recordID, "timeline.identity_refs", "identity", identityRefs, insertOptions, now)
	if err != nil {
		return mentionProjectionRefresh{}, err
	}
	if identityLinked {
		refresh.Identities = true
	}
	return refresh, nil
}

func applyCreateTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, tags *CollectionActionPayload, now time.Time) ([]recordTagMutation, error) {
	return insertTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, tags, now)
}

func (s *Store) applyPatchMentionActionsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) (mentionProjectionRefresh, error) {
	var refresh mentionProjectionRefresh
	for _, change := range changes {
		if change.ActionPayload == nil {
			continue
		}
		entityType := "host"
		if change.FieldKey == "timeline.identity_refs" {
			entityType = "identity"
		}
		for _, action := range change.ActionPayload.Actions {
			switch action.Op {
			case "add_token", "add_resolved_ref":
				linked, err := insertMentionActionsTx(ctx, tx, s.linkStore, actor.ID, incidentID, recordID, change.FieldKey, entityType, &CollectionActionPayload{Actions: []CollectionAction{action}}, mentionInsertOptions{
					allowInteractiveAutoResolution: true,
				}, now)
				if err != nil {
					return mentionProjectionRefresh{}, err
				}
				if linked {
					refresh.include(change.FieldKey)
				}
			case "resolve_item":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return mentionProjectionRefresh{}, err
				}
				if err := s.mentionStore.ResolveOrCreateFromMentionTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.ResolvedRecord, now); err != nil {
					return mentionProjectionRefresh{}, err
				}
			case "dismiss_item", "revert_to_unresolved":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return mentionProjectionRefresh{}, err
				}
				if err := s.mentionStore.ApplyMentionLifecycleTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.Op, nil, now); err != nil {
					return mentionProjectionRefresh{}, err
				}
			default:
				return mentionProjectionRefresh{}, fmt.Errorf("unsupported mention action: %s", action.Op)
			}
		}
	}
	return refresh, nil
}

func applyPatchTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) ([]recordTagMutation, error) {
	mutations := make([]recordTagMutation, 0)
	for _, change := range changes {
		if change.FieldKey != "timeline.tags" || change.ActionPayload == nil {
			continue
		}
		applied, err := insertTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func applyPatchAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) ([]attachedEvidenceMutation, error) {
	var mutations []attachedEvidenceMutation
	for _, change := range changes {
		if change.FieldKey != "timeline.attached_evidence_ids" || change.ActionPayload == nil {
			continue
		}
		applied, err := applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func applyAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) ([]attachedEvidenceMutation, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return nil, nil
	}
	mutations := make([]attachedEvidenceMutation, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, fmt.Errorf("missing linked evidence record")
			}
			if err := validateAttachedEvidenceTargetTx(ctx, tx, incidentID, *action.LinkedRecordID); err != nil {
				return nil, err
			}
			var linkID uuid.UUID
			err := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', NULL, $4, $4, $5, $5)
ON CONFLICT DO NOTHING
RETURNING record_link_id
`, incidentID, recordID, *action.LinkedRecordID, actorUserID, now.UTC()).Scan(&linkID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("insert attached evidence link: %w", err)
			}
			after, err := loadAttachedEvidenceLinkValueTx(ctx, tx, linkID)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, attachedEvidenceMutation{RecordLinkID: linkID, Operation: "create", AfterValue: after})
		case "remove_record_ref":
			evidenceRecordID, err := recordIDFromRecordRefItem(action.ItemRef)
			if err != nil {
				return nil, err
			}
			var linkID uuid.UUID
			err = tx.QueryRow(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, incidentID, recordID, evidenceRecordID).Scan(&linkID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("lookup attached evidence link: %w", err)
			}
			before, err := loadAttachedEvidenceLinkValueTx(ctx, tx, linkID)
			if err != nil {
				return nil, err
			}
			if _, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = COALESCE(deleted_at, $5),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $4)
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
`, incidentID, recordID, evidenceRecordID, actorUserID, now.UTC()); err != nil {
				return nil, fmt.Errorf("remove attached evidence link: %w", err)
			}
			after, err := loadAttachedEvidenceLinkValueTx(ctx, tx, linkID)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, attachedEvidenceMutation{RecordLinkID: linkID, Operation: "delete", BeforeValue: before, AfterValue: after})
		default:
			return nil, fmt.Errorf("unsupported attached evidence action: %s", action.Op)
		}
	}
	return mutations, nil
}

func (s *Store) insertAttachedEvidenceMutationEntriesTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequenceNo int, mutations []attachedEvidenceMutation) error {
	sequenceNo := startSequenceNo
	for _, mutation := range mutations {
		if mutation.RecordLinkID == uuid.Nil {
			continue
		}
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_link",
			TargetID:      mutation.RecordLinkID.String(),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
		sequenceNo++
	}
	return nil
}

func loadAttachedEvidenceLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_link_id', record_link_id::text,
    'incident_id', incident_id::text,
    'src_record_id', src_record_id::text,
    'dst_record_id', dst_record_id::text,
    'link_type', link_type,
    'field_key', field_key,
    'provenance', provenance,
    'confidence', confidence,
    'owner_user_id', owner_user_id::text,
    'created_by_user_id', created_by_user_id::text,
    'decided_at', decided_at,
    'created_at', created_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM record_links
 WHERE record_link_id = $1
`, recordLinkID).Scan(&raw); err != nil {
		return nil, fmt.Errorf("load attached evidence link value: %w", err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("decode attached evidence link value: %w", err)
	}
	return value, nil
}

func insertTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) ([]recordTagMutation, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return nil, nil
	}
	mutations := make([]recordTagMutation, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_tag":
			var tagID uuid.UUID
			err := tx.QueryRow(ctx, `
INSERT INTO record_tags (
    incident_id,
    record_id,
    tag_name,
    normalized_tag_name,
    created_by_user_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (incident_id, record_id, normalized_tag_name)
WHERE deleted_at IS NULL
DO NOTHING
RETURNING record_tag_id
`, incidentID, recordID, action.RawText, action.NormalizedText, actorUserID, now.UTC()).Scan(&tagID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("insert record tag: %w", err)
			}
			after, err := loadRecordTagValueTx(ctx, tx, tagID)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, recordTagMutation{RecordTagID: tagID, RecordID: recordID, Operation: "create", AfterValue: after})
		case "remove_tag":
			itemRecordID, tagID, err := recordTagItemRefParts(action.ItemRef)
			if err != nil || itemRecordID != recordID {
				return nil, fmt.Errorf("invalid record tag item ref")
			}
			before, err := loadRecordTagValueTx(ctx, tx, tagID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					continue
				}
				return nil, err
			}
			tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = $5,
       deleted_by_user_id = $4,
       updated_at = $5
 WHERE incident_id = $1
   AND record_id = $2
   AND record_tag_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, tagID, actorUserID, now.UTC())
			if err != nil {
				return nil, fmt.Errorf("remove record tag: %w", err)
			}
			if tag.RowsAffected() == 0 {
				continue
			}
			after, err := loadRecordTagValueTx(ctx, tx, tagID)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, recordTagMutation{RecordTagID: tagID, RecordID: recordID, Operation: "delete", BeforeValue: before, AfterValue: after})
		default:
			return nil, fmt.Errorf("unsupported tag action: %s", action.Op)
		}
	}
	return mutations, nil
}

func (s *Store) insertRecordTagMutationEntriesTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequenceNo int, mutations []recordTagMutation) error {
	sequenceNo := startSequenceNo
	for _, mutation := range mutations {
		if mutation.RecordTagID == uuid.Nil || mutation.RecordID == uuid.Nil {
			continue
		}
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_tag",
			TargetID:      recordTagMutationTarget(mutation.RecordID, mutation.RecordTagID),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
		sequenceNo++
	}
	return nil
}

func loadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_tag_id', record_tag_id::text,
    'tag_id', record_tag_id::text,
    'incident_id', incident_id::text,
    'record_id', record_id::text,
    'tag_name', tag_name,
    'normalized_tag_name', normalized_tag_name,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'updated_at', updated_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM record_tags
 WHERE record_tag_id = $1
`, recordTagID).Scan(&raw); err != nil {
		return nil, fmt.Errorf("load record tag value: %w", err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("decode record tag value: %w", err)
	}
	return value, nil
}

func recordTagMutationTarget(recordID uuid.UUID, tagID uuid.UUID) string {
	return "record_tag:" + recordID.String() + ":" + tagID.String()
}

func recordTagItemRefParts(itemRef string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	tagID, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	return recordID, tagID, nil
}

func insertMentionActionsTx(ctx context.Context, tx pgx.Tx, linkStore timelineLinkPort, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, fieldKey string, entityType string, payload *CollectionActionPayload, options mentionInsertOptions, now time.Time) (bool, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return false, nil
	}
	originKind := options.originKind
	if strings.TrimSpace(originKind) == "" {
		originKind = "interactive_cell"
	}
	nextOrdinal, err := nextMentionOrdinalTx(ctx, tx, recordID, fieldKey)
	if err != nil {
		return false, err
	}
	linked := false
	for _, action := range payload.Actions {
		if action.Op != "add_token" && action.Op != "add_resolved_ref" {
			return false, fmt.Errorf("unsupported mention action: %s", action.Op)
		}
		resolutionStatus := "unresolved"
		var resolvedRecordID *uuid.UUID
		var resolvedByUserID any
		var resolvedAt any
		var resolutionMethod any
		linkProvenance := "manual"
		var linkConfidence *int
		if action.Op == "add_resolved_ref" {
			if action.ResolvedRecord == nil {
				return false, fmt.Errorf("missing resolved record for action: %s", action.Op)
			}
			if err := validateTimelineResolvedTargetTx(ctx, tx, incidentID, entityType, *action.ResolvedRecord); err != nil {
				return false, err
			}
			resolutionStatus = "resolved"
			resolvedRecordID = action.ResolvedRecord
			resolvedByUserID = actorUserID
			resolvedAt = now.UTC()
			resolutionMethod = action.Op
		} else if options.allowInteractiveAutoResolution {
			match, err := lookupInteractiveAutoResolutionMatchTx(ctx, tx, incidentID, fieldKey, action.RawText)
			if err != nil {
				return false, err
			}
			if match != nil {
				resolutionStatus = "resolved"
				resolvedRecordID = &match.RecordID
				resolvedByUserID = actorUserID
				resolvedAt = now.UTC()
				resolutionMethod = autoResolutionMethod
				linkProvenance = autoResolutionMethod
				confidence := 100
				linkConfidence = &confidence
			}
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO entity_mentions (
    source_record_id,
    entity_type,
    source_field_key,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    ordinal,
    created_by_user_id,
    created_at,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, $9, $10, $11, $12, $13, $14, $15)
`, recordID, entityType, fieldKey, originKind, mentionOriginLocator(recordID, fieldKey, nextOrdinal), action.RawText, action.NormalizedText, resolutionStatus, nextOrdinal, actorUserID, now.UTC(), resolvedRecordID, resolvedByUserID, resolvedAt, resolutionMethod); err != nil {
			return false, fmt.Errorf("insert entity mention: %w", err)
		}
		if resolvedRecordID != nil {
			linkType, ok := timelineRelationshipLinkType(fieldKey)
			if !ok {
				return false, fmt.Errorf("unsupported link field: %s", fieldKey)
			}
			if err := linkStore.UpsertLinkTx(ctx, tx, incidentID, recordID, *resolvedRecordID, linkType, linkProvenance, linkConfidence, actorUserID, now.UTC()); err != nil {
				return false, fmt.Errorf("upsert record link: %w", err)
			}
			linked = true
		}
		nextOrdinal++
	}
	return linked, nil
}

func timelineRelationshipLinkType(fieldKey string) (string, bool) {
	switch fieldKey {
	case "timeline.host_refs":
		return "observed_on_host", true
	case "timeline.identity_refs":
		return "observed_as_identity", true
	default:
		return "", false
	}
}

func nextMentionOrdinalTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string) (int, error) {
	var nextOrdinal int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(ordinal), 0) + 1
  FROM entity_mentions
 WHERE source_record_id = $1
   AND source_field_key = $2
`, recordID, fieldKey).Scan(&nextOrdinal); err != nil {
		return 0, fmt.Errorf("query next mention ordinal: %w", err)
	}
	return nextOrdinal, nil
}

func validateTimelineResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, resolvedRecordID uuid.UUID) error {
	var exists bool
	switch entityType {
	case "host":
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM hosts
     WHERE record_id = $1
       AND incident_id = $2
       AND host_state IN ('stub', 'canonical')
)
`, resolvedRecordID, incidentID).Scan(&exists); err != nil {
			return fmt.Errorf("validate host target: %w", err)
		}
	case "identity":
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM identities
     WHERE record_id = $1
       AND incident_id = $2
       AND identity_state IN ('stub', 'canonical')
)
`, resolvedRecordID, incidentID).Scan(&exists); err != nil {
			return fmt.Errorf("validate identity target: %w", err)
		}
	default:
		return entities.ErrResolvedRecordNotFound
	}
	if !exists {
		return entities.ErrResolvedRecordNotFound
	}
	return nil
}

func validateAttachedEvidenceTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records r
      JOIN evidence ev
        ON ev.incident_id = r.incident_id
       AND ev.record_id = r.record_id
     WHERE r.record_id = $1
       AND r.incident_id = $2
       AND r.record_type = 'evidence'
       AND r.deleted_at IS NULL
)
`, evidenceRecordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate attached evidence target: %w", err)
	}
	if !exists {
		return entities.ErrResolvedRecordNotFound
	}
	return nil
}

func recordIDFromRecordRefItem(itemRef string) (uuid.UUID, error) {
	const prefix = "record_ref:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid record ref item_ref: %s", itemRef)
	}
	recordID, err := uuid.Parse(strings.TrimPrefix(itemRef, prefix))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse record ref item_ref: %w", err)
	}
	return recordID, nil
}

func mentionOriginLocator(recordID uuid.UUID, fieldKey string, ordinal int) string {
	return fmt.Sprintf("view:%s/record:%s/cell:%s/item:%d", TimelineViewSchemaID, recordID.String(), fieldKey, ordinal)
}

func mentionIDFromItemRef(itemRef string) (uuid.UUID, error) {
	const prefix = "entity_mention:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid mention item_ref: %s", itemRef)
	}
	mentionID, err := uuid.Parse(strings.TrimPrefix(itemRef, prefix))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse mention item_ref: %w", err)
	}
	return mentionID, nil
}
