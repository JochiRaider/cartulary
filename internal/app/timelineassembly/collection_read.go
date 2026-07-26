package timelineassembly

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

const autoResolutionMethod = "auto_match"

type timelineProjectedRecord = workbookprojection.DerivedRecord

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

func hydrateTimelineCollections(ctx context.Context, querier mentionQueryer, record *timelineProjectedRecord) error {
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
			if linkType, ok := mentionLinkType(mention.SourceFieldKey); ok {
				linkMetadata, err := loadActiveTimelineCollectionLinkMetadata(ctx, querier, record.IncidentID, record.RecordID, *mention.ResolvedRecordID, linkType)
				if err != nil {
					return err
				}
				if linkMetadata != nil {
					item["provenance"] = linkMetadata.Provenance
					item["confidence"] = linkMetadata.Confidence
				}
			}
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod == autoResolutionMethod {
				matchedAliasText, err := loadMatchedTimelineAliasText(ctx, querier, *mention.ResolvedRecordID, mention.EntityType, mention.RawText)
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
  FROM active_record_tags_v1
 WHERE incident_id = $1
   AND record_id = $2
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
			"item_ref":     links.RecordTagItemRef(record.RecordID, recordTagID),
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
  FROM active_record_links_v1 rl
  JOIN evidence ev
    ON ev.incident_id = rl.incident_id
   AND ev.record_id = rl.dst_record_id
  LEFT JOIN object_blobs b
    ON b.object_blob_id = ev.object_blob_id
WHERE rl.incident_id = $1
   AND rl.src_record_id = $2
   AND rl.link_type = 'attached_evidence'
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
			"item_ref":         links.RecordRefItemRef(evidenceRecordID),
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

	replacementRows, err := querier.Query(ctx, `
SELECT src_record_id
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline replacement link: %w", err)
	}
	if replacementRows.Next() {
		var replacementRecordID uuid.UUID
		if err := replacementRows.Scan(&replacementRecordID); err != nil {
			replacementRows.Close()
			return fmt.Errorf("scan timeline replacement link: %w", err)
		}
		record.ReplacementRecordID = &replacementRecordID
	}
	if err := replacementRows.Err(); err != nil {
		replacementRows.Close()
		return fmt.Errorf("iterate timeline replacement links: %w", err)
	}
	replacementRows.Close()
	return nil
}

type timelineCollectionLinkMetadata struct {
	Provenance string
	Confidence *int
}

func loadActiveTimelineCollectionLinkMetadata(ctx context.Context, querier mentionQueryer, incidentID uuid.UUID, sourceRecordID uuid.UUID, targetRecordID uuid.UUID, linkType string) (*timelineCollectionLinkMetadata, error) {
	rows, err := querier.Query(ctx, `
SELECT provenance, confidence
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, incidentID, sourceRecordID, targetRecordID, linkType)
	if err != nil {
		return nil, fmt.Errorf("query active timeline collection link metadata: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate active timeline collection link metadata: %w", err)
		}
		return nil, nil
	}

	var (
		metadata   timelineCollectionLinkMetadata
		confidence pgtype.Int4
	)
	if err := rows.Scan(&metadata.Provenance, &confidence); err != nil {
		return nil, fmt.Errorf("scan active timeline collection link metadata: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		metadata.Confidence = &value
	}
	return &metadata, nil
}

func loadMatchedTimelineAliasText(ctx context.Context, querier mentionQueryer, recordID uuid.UUID, entityType string, rawText string) (*string, error) {
	candidateText, ok := fieldnorm.AutoResolutionCandidateText(rawText)
	if !ok {
		return nil, nil
	}

	rows, err := querier.Query(ctx, `
SELECT raw_text
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY created_at ASC, entity_alias_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("query matched timeline alias text: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var aliasText string
		if err := rows.Scan(&aliasText); err != nil {
			return nil, fmt.Errorf("scan matched timeline alias text: %w", err)
		}
		aliasCandidateText, ok := fieldnorm.AutoResolutionCandidateText(aliasText)
		if ok && aliasCandidateText == candidateText {
			value := aliasText
			return &value, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matched timeline alias texts: %w", err)
	}
	return nil, nil
}

func mentionLinkType(sourceFieldKey string) (string, bool) {
	switch sourceFieldKey {
	case "timeline.host_refs":
		return "observed_on_host", true
	case "timeline.identity_refs":
		return "observed_as_identity", true
	default:
		return "", false
	}
}
