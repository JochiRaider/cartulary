package entities

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrEntityMentionNotFound = errors.New("entities: entity mention not found")
var ErrResolvedRecordNotFound = errors.New("entities: resolved record not found")
var ErrSourceRecordNotFound = errors.New("entities: source record not found")
var ErrRecordDeletedUseRestore = errors.New("entities: record deleted use restore")

type MentionTransitionError struct {
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
}

func (e *MentionTransitionError) Error() string {
	return fmt.Sprintf("entities: illegal mention transition %s -> %s", e.FromStatus, e.ToStatus)
}

type MentionRowVersionConflictError struct {
	EntityMentionID          uuid.UUID
	BaseMentionRowVersion    int64
	CurrentMentionRowVersion int64
	SourceRecordID           uuid.UUID
}

func (e *MentionRowVersionConflictError) Error() string {
	return fmt.Sprintf("entities: mention row version conflict for %s", e.EntityMentionID)
}

type MentionTargetValidationError struct {
	Reason string
}

func (e *MentionTargetValidationError) Error() string {
	return fmt.Sprintf("entities: invalid resolved target: %s", e.Reason)
}

type mentionActionRecord struct {
	EntityMentionID  uuid.UUID
	IncidentID       uuid.UUID
	SourceRecordID   uuid.UUID
	SourceFieldKey   string
	EntityType       string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
	Role             string
}

type mentionTargetRecord struct {
	RecordID   uuid.UUID
	IncidentID uuid.UUID
	EntityType string
}

type mentionMutationResult struct {
	Before         mentionActionRecord
	After          mentionActionRecord
	ActiveLink     *links.RecordLink
	CreatedLink    *links.RecordLink
	TombstonedLink *links.RecordLink
}

func (s *Store) GetMentionActionAccess(ctx context.Context, mentionID uuid.UUID, userID uuid.UUID) (mentionActionRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT
    m.entity_mention_id,
    r.incident_id,
    m.source_record_id,
    m.source_field_key,
    m.entity_type,
    m.origin_kind,
    m.origin_locator,
    m.raw_text,
    m.normalized_text,
    m.resolution_status,
    m.row_version,
    m.resolved_record_id,
    m.resolved_by_user_id,
    m.resolved_at,
    m.resolution_method,
    im.role
  FROM entity_mentions m
  JOIN records r ON r.record_id = m.source_record_id
  LEFT JOIN incident_memberships im
    ON im.incident_id = r.incident_id
   AND im.user_id = $2
 WHERE m.entity_mention_id = $1
`, mentionID, userID)
	record, err := scanMentionActionRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return mentionActionRecord{}, ErrEntityMentionNotFound
	}
	if err != nil {
		return mentionActionRecord{}, fmt.Errorf("load mention action access: %w", err)
	}
	if strings.TrimSpace(record.Role) == "" {
		return mentionActionRecord{}, ErrEntityMentionNotFound
	}
	return record, nil
}

func (s *Store) ApplyMentionAction(ctx context.Context, actor authn.UserRecord, mentionID uuid.UUID, request MentionActionRequest, requestHash []byte, requestID string, now time.Time) (MentionActionResult, error) {
	scopeKey := mentionIdempotencyScope(actor.ID, mentionID)
	if existing, err := s.authStore.GetRouteIdempotency(ctx, mentionActionRouteKey, scopeKey, request.ClientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MentionActionResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MentionActionResult{}, fmt.Errorf("decode replayed mention action payload: %w", err)
		}
		return MentionActionResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MentionActionResult{}, fmt.Errorf("query mention action idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MentionActionResult{}, fmt.Errorf("begin mention action transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	mention, err := loadMentionActionRecordTx(ctx, tx, mentionID)
	if err != nil {
		return MentionActionResult{}, err
	}
	if mention.RowVersion != request.BaseMentionRowVersion {
		return MentionActionResult{}, &MentionRowVersionConflictError{
			EntityMentionID:          mention.EntityMentionID,
			BaseMentionRowVersion:    request.BaseMentionRowVersion,
			CurrentMentionRowVersion: mention.RowVersion,
			SourceRecordID:           mention.SourceRecordID,
		}
	}

	sourceRecord, err := loadTimelineSourceRecordTx(ctx, tx, mention.SourceRecordID)
	if err != nil {
		return MentionActionResult{}, err
	}
	beforeProjected := projectTimelineRecord(sourceRecord, nil)
	if err := hydrateTimelineCollections(ctx, tx, &beforeProjected); err != nil {
		return MentionActionResult{}, err
	}

	var validatedTarget *mentionTargetRecord
	if request.Action == "resolve_item" {
		validatedTarget, err = validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *request.ResolvedRecordID)
		if err != nil {
			return MentionActionResult{}, err
		}
	}

	outcome, err := s.applyMentionActionTx(ctx, tx, actor.ID, mention, request.Action, validatedTarget, resolutionMethodPointer("explicit_resolve_route"), now.UTC())
	if err != nil {
		return MentionActionResult{}, err
	}

	nextRecord := sourceRecord
	nextRecord.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, sourceRecord.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MentionActionResult{}, err
	}
	nextRecord.EditedAt = now.UTC()
	nextRecord.UpdatedByUserID = actor.ID
	if err := updateMentionSourceRecordTx(ctx, tx, nextRecord); err != nil {
		return MentionActionResult{}, err
	}

	afterProjected := projectTimelineRecord(nextRecord, nil)
	if err := hydrateTimelineCollections(ctx, tx, &afterProjected); err != nil {
		return MentionActionResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  mention.IncidentID,
		ActorUserID: actor.ID,
		Source:      mentionActionRouteKey,
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MentionActionResult{}, err
	}

	beforeRow := buildTimelineRow(beforeProjected)
	afterRow := buildTimelineRow(afterProjected)
	beforeVersionID := timelineVersionID(sourceRecord.RecordID, sourceRecord.RowVersion)
	afterVersionID := timelineVersionID(nextRecord.RecordID, nextRecord.RowVersion)
	sequenceNo := 1
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      sequenceNo,
		TargetKind:      "timeline_record",
		TargetID:        sourceRecord.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MentionActionResult{}, err
	}
	sequenceNo++
	beforeMentionVersion := mentionVersionID(outcome.Before.EntityMentionID, outcome.Before.RowVersion)
	afterMentionVersion := mentionVersionID(outcome.After.EntityMentionID, outcome.After.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      sequenceNo,
		TargetKind:      "entity_mention",
		TargetID:        outcome.After.EntityMentionID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeMentionVersion,
		AfterVersionID:  &afterMentionVersion,
		BeforeValue:     buildMentionMutationValue(outcome.Before),
		AfterValue:      buildMentionMutationValue(outcome.After),
	}); err != nil {
		return MentionActionResult{}, err
	}
	sequenceNo++
	if outcome.TombstonedLink != nil {
		beforeLink := buildLinkMutationValue(*outcome.TombstonedLink, nil)
		afterLink := buildLinkMutationValue(*outcome.TombstonedLink, outcome.TombstonedLink.DeletedAt)
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_link",
			TargetID:      outcome.TombstonedLink.RecordLinkID.String(),
			OperationKind: "delete",
			BeforeValue:   beforeLink,
			AfterValue:    afterLink,
		}); err != nil {
			return MentionActionResult{}, err
		}
		sequenceNo++
	}
	if outcome.CreatedLink != nil {
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_link",
			TargetID:      outcome.CreatedLink.RecordLinkID.String(),
			OperationKind: "create",
			AfterValue:    buildLinkMutationValue(*outcome.CreatedLink, nil),
		}); err != nil {
			return MentionActionResult{}, err
		}
		sequenceNo++
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    sourceRecord.RecordID,
		RowVersion:  nextRecord.RowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MentionActionResult{}, err
	}

	if err := s.projectionStore.RebuildIncidentTimelineTx(ctx, tx, mention.IncidentID); err != nil {
		return MentionActionResult{}, err
	}

	payload := buildMentionActionPayload(mention.IncidentID, outcome.After, nextRecord.RecordID, nextRecord.RowVersion, changeSetID, outcome.ActiveLink)
	if err := insertRouteIdempotency(ctx, tx, mentionActionRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MentionActionResult{}, authn.ErrClientTxnConflict
		}
		return MentionActionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MentionActionResult{}, fmt.Errorf("commit mention action transaction: %w", err)
	}

	return MentionActionResult{
		Payload:                payload,
		StatusCode:             http.StatusOK,
		IncidentID:             mention.IncidentID,
		SourceRecordID:         nextRecord.RecordID,
		SourceRecordRowVersion: nextRecord.RowVersion,
		ChangeSetID:            changeSetID,
		ChangedFieldKeys:       mentionChangedFieldKeys(mention.SourceFieldKey),
	}, nil
}

func (s *Store) ApplyMentionLifecycleTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, sourceFieldKey string, mentionID uuid.UUID, action string, resolvedRecordID *uuid.UUID, now time.Time) error {
	mention, err := loadMentionActionRecordTx(ctx, tx, mentionID)
	if err != nil {
		return err
	}
	if mention.SourceRecordID != sourceRecordID || mention.SourceFieldKey != sourceFieldKey {
		return ErrInvalidMentionResolution
	}

	var validatedTarget *mentionTargetRecord
	if action == "resolve_item" {
		if resolvedRecordID == nil {
			return ErrInvalidMentionResolution
		}
		validatedTarget, err = validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *resolvedRecordID)
		if err != nil {
			return err
		}
	}
	if _, err := s.applyMentionActionTx(ctx, tx, actor.ID, mention, action, validatedTarget, resolutionMethodPointer("explicit_resolve_route"), now.UTC()); err != nil {
		return err
	}
	return nil
}

func (s *Store) applyMentionActionTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, mention mentionActionRecord, action string, target *mentionTargetRecord, resolutionMethod *string, now time.Time) (mentionMutationResult, error) {
	linkType, ok := mentionLinkType(mention.SourceFieldKey)
	toStatus, legalStatuses := mentionActionState(action)
	guards := make([]string, 0, 2)
	if !ok {
		guards = append(guards, "source_field_not_supported")
	}
	if !slices.Contains(legalStatuses, mention.ResolutionStatus) {
		guards = append(guards, "from_status_not_allowed")
	}
	if len(guards) > 0 {
		return mentionMutationResult{}, &MentionTransitionError{
			FromStatus:     mention.ResolutionStatus,
			ToStatus:       toStatus,
			ViolatedGuards: guards,
		}
	}

	before := mention
	after := mention
	after.RowVersion++
	after.ResolutionStatus = toStatus

	switch action {
	case "resolve_item":
		if target == nil {
			return mentionMutationResult{}, &MentionTargetValidationError{Reason: "missing_target"}
		}
		targetID := target.RecordID
		after.ResolvedRecordID = &targetID
		resolvedBy := actorUserID
		after.ResolvedByUserID = &resolvedBy
		resolvedAt := now.UTC()
		after.ResolvedAt = &resolvedAt
		after.ResolutionMethod = cloneStringPointer(resolutionMethod)
	case "dismiss_item", "revert_to_unresolved":
		after.ResolvedRecordID = nil
		after.ResolvedByUserID = nil
		after.ResolvedAt = nil
		after.ResolutionMethod = nil
	default:
		return mentionMutationResult{}, &MentionTargetValidationError{Reason: "unsupported_action"}
	}

	if _, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET resolution_status = $2,
       row_version = $3,
       resolved_record_id = $4,
       resolved_by_user_id = $5,
       resolved_at = $6,
       resolution_method = $7
 WHERE entity_mention_id = $1
`, mention.EntityMentionID, after.ResolutionStatus, after.RowVersion, after.ResolvedRecordID, after.ResolvedByUserID, after.ResolvedAt, after.ResolutionMethod); err != nil {
		return mentionMutationResult{}, fmt.Errorf("update entity mention lifecycle: %w", err)
	}

	var (
		activeLink     *links.RecordLink
		createdLink    *links.RecordLink
		tombstonedLink *links.RecordLink
	)
	if before.ResolvedRecordID != nil && (after.ResolvedRecordID == nil || *after.ResolvedRecordID != *before.ResolvedRecordID) {
		removeOldLink, err := shouldTombstoneMentionLinkTx(ctx, tx, before, *before.ResolvedRecordID)
		if err != nil {
			return mentionMutationResult{}, err
		}
		if removeOldLink {
			existingLink, err := s.linkStore.GetActiveLinkTx(ctx, tx, before.IncidentID, before.SourceRecordID, *before.ResolvedRecordID, linkType)
			switch {
			case errors.Is(err, links.ErrRecordLinkNotFound):
			case err != nil:
				return mentionMutationResult{}, err
			default:
				tombstoned, err := s.linkStore.TombstoneLinkTx(ctx, tx, existingLink.RecordLinkID, actorUserID, now.UTC())
				if err != nil {
					return mentionMutationResult{}, err
				}
				tombstonedLink = &tombstoned
			}
		}
	}
	if after.ResolvedRecordID != nil {
		link, inserted, err := s.linkStore.UpsertLinkTx(ctx, tx, after.IncidentID, after.SourceRecordID, *after.ResolvedRecordID, linkType, "manual", nil, actorUserID, now.UTC())
		if err != nil {
			return mentionMutationResult{}, err
		}
		activeLink = &link
		if inserted {
			createdLink = &link
		}
	}

	return mentionMutationResult{
		Before:         before,
		After:          after,
		ActiveLink:     activeLink,
		CreatedLink:    createdLink,
		TombstonedLink: tombstonedLink,
	}, nil
}

func loadMentionActionRecordTx(ctx context.Context, tx pgx.Tx, mentionID uuid.UUID) (mentionActionRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    m.entity_mention_id,
    r.incident_id,
    m.source_record_id,
    m.source_field_key,
    m.entity_type,
    m.origin_kind,
    m.origin_locator,
    m.raw_text,
    m.normalized_text,
    m.resolution_status,
    m.row_version,
    m.resolved_record_id,
    m.resolved_by_user_id,
    m.resolved_at,
    m.resolution_method
    ,
    NULL::text AS role
  FROM entity_mentions m
  JOIN records r ON r.record_id = m.source_record_id
 WHERE m.entity_mention_id = $1
 FOR UPDATE OF m, r
`, mentionID)
	record, err := scanMentionActionRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return mentionActionRecord{}, ErrEntityMentionNotFound
	}
	if err != nil {
		return mentionActionRecord{}, fmt.Errorf("load mention action record: %w", err)
	}
	return record, nil
}

func validateMentionResolvedTargetTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, entityType string, resolvedRecordID uuid.UUID) (*mentionTargetRecord, error) {
	rows, err := tx.Query(ctx, `
SELECT entity_type, incident_id, record_id
  FROM (
        SELECT 'host'::text AS entity_type, h.incident_id, h.record_id
          FROM hosts h
          JOIN incident_memberships im
            ON im.incident_id = h.incident_id
           AND im.user_id = $1
         WHERE h.record_id = $2
           AND h.host_state IN ('stub', 'canonical')
        UNION ALL
        SELECT 'identity'::text AS entity_type, i.incident_id, i.record_id
          FROM identities i
          JOIN incident_memberships im
            ON im.incident_id = i.incident_id
           AND im.user_id = $1
         WHERE i.record_id = $2
           AND i.identity_state IN ('stub', 'canonical')
  ) visible_records
`, actorUserID, resolvedRecordID)
	if err != nil {
		return nil, fmt.Errorf("query mention resolved target: %w", err)
	}
	defer rows.Close()

	var target *mentionTargetRecord
	for rows.Next() {
		var record mentionTargetRecord
		if err := rows.Scan(&record.EntityType, &record.IncidentID, &record.RecordID); err != nil {
			return nil, fmt.Errorf("scan mention resolved target: %w", err)
		}
		target = &record
		break
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mention resolved target: %w", err)
	}
	if target == nil {
		return nil, ErrResolvedRecordNotFound
	}
	if target.EntityType != entityType || target.IncidentID != incidentID {
		return nil, &MentionTargetValidationError{Reason: "incompatible_target"}
	}
	return target, nil
}

func shouldTombstoneMentionLinkTx(ctx context.Context, tx pgx.Tx, mention mentionActionRecord, resolvedRecordID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM entity_mentions
     WHERE source_record_id = $1
       AND source_field_key = $2
       AND entity_mention_id <> $3
       AND resolution_status = 'resolved'
       AND resolved_record_id = $4
)
`, mention.SourceRecordID, mention.SourceFieldKey, mention.EntityMentionID, resolvedRecordID).Scan(&exists); err != nil {
		return false, fmt.Errorf("query sibling resolved mentions: %w", err)
	}
	return !exists, nil
}

func updateMentionSourceRecordTx(ctx context.Context, tx pgx.Tx, record timelineSourceRecord) error {
	if _, err := tx.Exec(ctx, `
UPDATE timeline_events
   SET row_version = $2,
       edited_at = $3,
       updated_by_user_id = $4
 WHERE record_id = $1
`, record.RecordID, record.RowVersion, record.EditedAt.UTC(), record.UpdatedByUserID); err != nil {
		return fmt.Errorf("update mention source record: %w", err)
	}
	return nil
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

func mentionActionState(action string) (string, []string) {
	switch action {
	case "resolve_item":
		return "resolved", []string{"unresolved", "resolved"}
	case "dismiss_item":
		return "dismissed", []string{"unresolved", "resolved"}
	case "revert_to_unresolved":
		return "unresolved", []string{"resolved", "dismissed"}
	default:
		return "", nil
	}
}

func buildMentionActionPayload(incidentID uuid.UUID, mention mentionActionRecord, sourceRecordID uuid.UUID, sourceRecordRowVersion int64, changeSetID uuid.UUID, activeLink *links.RecordLink) map[string]any {
	data := map[string]any{
		"incident_id": incidentID.String(),
		"entity_mention": map[string]any{
			"entity_mention_id":   mention.EntityMentionID.String(),
			"source_record_id":    mention.SourceRecordID.String(),
			"source_field_key":    mention.SourceFieldKey,
			"entity_type":         mention.EntityType,
			"raw_text":            mention.RawText,
			"normalized_text":     mention.NormalizedText,
			"resolution_status":   mention.ResolutionStatus,
			"resolved_record_id":  formatUUIDPointer(mention.ResolvedRecordID),
			"row_version":         mention.RowVersion,
			"resolved_at":         formatTimestampPointer(mention.ResolvedAt),
			"resolved_by_user_id": formatUUIDPointer(mention.ResolvedByUserID),
			"resolution_method":   derefString(mention.ResolutionMethod),
		},
		"source_record": map[string]any{
			"record_id":   sourceRecordID.String(),
			"row_version": sourceRecordRowVersion,
		},
		"change_set_id": changeSetID.String(),
	}
	if activeLink != nil && activeLink.DeletedAt == nil {
		data["active_link"] = map[string]any{
			"record_link_id": activeLink.RecordLinkID.String(),
			"src_record_id":  activeLink.SrcRecordID.String(),
			"dst_record_id":  activeLink.DstRecordID.String(),
			"link_type":      activeLink.LinkType,
		}
	}
	return data
}

func buildMentionMutationValue(mention mentionActionRecord) map[string]any {
	return map[string]any{
		"entity_mention_id":   mention.EntityMentionID.String(),
		"incident_id":         mention.IncidentID.String(),
		"source_record_id":    mention.SourceRecordID.String(),
		"source_field_key":    mention.SourceFieldKey,
		"entity_type":         mention.EntityType,
		"origin_kind":         mention.OriginKind,
		"origin_locator":      mention.OriginLocator,
		"raw_text":            mention.RawText,
		"normalized_text":     mention.NormalizedText,
		"resolution_status":   mention.ResolutionStatus,
		"row_version":         mention.RowVersion,
		"resolved_record_id":  formatUUIDPointer(mention.ResolvedRecordID),
		"resolved_by_user_id": formatUUIDPointer(mention.ResolvedByUserID),
		"resolved_at":         formatTimestampPointer(mention.ResolvedAt),
		"resolution_method":   derefString(mention.ResolutionMethod),
	}
}

func buildLinkMutationValue(link links.RecordLink, deletedAt *time.Time) map[string]any {
	value := map[string]any{
		"record_link_id": link.RecordLinkID.String(),
		"incident_id":    link.IncidentID.String(),
		"src_record_id":  link.SrcRecordID.String(),
		"dst_record_id":  link.DstRecordID.String(),
		"link_type":      link.LinkType,
		"provenance":     link.Provenance,
		"confidence":     link.Confidence,
		"deleted_at":     formatTimestampPointer(deletedAt),
	}
	return value
}

func mentionVersionID(mentionID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("entity_mention:%s:%d", mentionID.String(), rowVersion)
}

func timelineVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("timeline_record:%s:%d", recordID.String(), rowVersion)
}

func mentionChangedFieldKeys(sourceFieldKey string) []string {
	keys := []string{sourceFieldKey, "timeline.has_unresolved_mentions"}
	slices.Sort(keys)
	return keys
}

func mentionIdempotencyScope(actorUserID uuid.UUID, mentionID uuid.UUID) string {
	return actorUserID.String() + ":" + mentionID.String()
}

func resolutionMethodPointer(value string) *string {
	return &value
}

func scanMentionActionRecord(row pgx.Row) (mentionActionRecord, error) {
	var (
		record           mentionActionRecord
		resolvedRecordID pgtype.UUID
		resolvedByUserID pgtype.UUID
		resolvedAt       pgtype.Timestamptz
		resolutionMethod pgtype.Text
		role             pgtype.Text
	)
	if err := row.Scan(
		&record.EntityMentionID,
		&record.IncidentID,
		&record.SourceRecordID,
		&record.SourceFieldKey,
		&record.EntityType,
		&record.OriginKind,
		&record.OriginLocator,
		&record.RawText,
		&record.NormalizedText,
		&record.ResolutionStatus,
		&record.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
		&role,
	); err != nil {
		return mentionActionRecord{}, err
	}
	if resolvedRecordID.Valid {
		value := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
		record.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := uuid.Must(uuid.FromBytes(resolvedByUserID.Bytes[:]))
		record.ResolvedByUserID = &value
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		record.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		record.ResolutionMethod = &value
	}
	if role.Valid {
		record.Role = role.String
	}
	return record, nil
}
