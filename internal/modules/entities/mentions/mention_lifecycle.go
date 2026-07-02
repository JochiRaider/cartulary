package mentions

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

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
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
	ActiveLink     *recordLink
	CreatedLink    *recordLink
	TombstonedLink *recordLink
}

type MentionActionAccess struct {
	Role string
}

func (s *Store) GetMentionActionAccess(ctx context.Context, mentionID uuid.UUID, userID uuid.UUID) (MentionActionAccess, error) {
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
		return MentionActionAccess{}, ErrEntityMentionNotFound
	}
	if err != nil {
		return MentionActionAccess{}, fmt.Errorf("load mention action access: %w", err)
	}
	if strings.TrimSpace(record.Role) == "" {
		return MentionActionAccess{}, ErrEntityMentionNotFound
	}
	return MentionActionAccess{Role: record.Role}, nil
}

func (s *Store) ApplyMentionAction(ctx context.Context, actor authn.UserRecord, mentionID uuid.UUID, request MentionActionRequest, requestHash []byte, requestID string, now time.Time) (MentionActionResult, error) {
	scopeKey := mentionIdempotencyScope(mentionID)
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    mentionActionRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
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
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, mention.IncidentID); err != nil {
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

	timelineState, err := s.ports.timeline.PrepareMentionActionTx(ctx, tx, mention.SourceRecordID)
	if err != nil {
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

	timelineResult, err := s.ports.timeline.ApplyMentionActionEffectsTx(ctx, tx, timelineState, timelineMentionActionCommand{
		IncidentID:  mention.IncidentID,
		ActorUserID: actor.ID,
		EffectiveAt: now.UTC(),
	})
	if err != nil {
		return MentionActionResult{}, err
	}

	changeSetID, err := s.ports.revisions.InsertChangeSetTx(ctx, tx, changeSetParams{
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

	sequenceNo := 1
	if err := s.ports.revisions.InsertMutationTx(ctx, tx, mutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      sequenceNo,
		TargetKind:      "timeline_record",
		TargetID:        timelineResult.SourceRecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &timelineResult.BeforeVersionID,
		AfterVersionID:  &timelineResult.AfterVersionID,
		BeforeValue:     timelineResult.BeforeRow,
		AfterValue:      timelineResult.AfterRow,
	}); err != nil {
		return MentionActionResult{}, err
	}
	sequenceNo++
	beforeMentionVersion := mentionVersionID(outcome.Before.EntityMentionID, outcome.Before.RowVersion)
	afterMentionVersion := mentionVersionID(outcome.After.EntityMentionID, outcome.After.RowVersion)
	if err := s.ports.revisions.InsertMutationTx(ctx, tx, mutationParams{
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
		if err := s.ports.revisions.InsertMutationTx(ctx, tx, mutationParams{
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
		if err := s.ports.revisions.InsertMutationTx(ctx, tx, mutationParams{
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
	if err := s.ports.revisions.InsertRecordRevisionTx(ctx, tx, recordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    timelineResult.SourceRecordID,
		RowVersion:  timelineResult.RowVersion,
		BeforeValue: timelineResult.BeforeRow,
		AfterValue:  timelineResult.AfterRow,
	}); err != nil {
		return MentionActionResult{}, err
	}

	entityInvalidations, err := s.mentionEntityInvalidationsTx(ctx, tx, outcome)
	if err != nil {
		return MentionActionResult{}, err
	}

	payload := buildMentionActionPayload(mention.IncidentID, outcome.After, timelineResult.SourceRecordID, timelineResult.RowVersion, changeSetID, outcome.ActiveLink)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
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
		SourceRecordID:         timelineResult.SourceRecordID,
		SourceRecordRowVersion: timelineResult.RowVersion,
		ChangeSetID:            changeSetID,
		ClientTxnID:            request.ClientTxnID,
		ChangedFieldKeys:       mentionChangedFieldKeys(mention.SourceFieldKey),
		EntityInvalidations:    entityInvalidations,
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
		activeLink     *recordLink
		createdLink    *recordLink
		tombstonedLink *recordLink
	)
	if before.ResolvedRecordID != nil && (after.ResolvedRecordID == nil || *after.ResolvedRecordID != *before.ResolvedRecordID) {
		removeOldLink, err := shouldTombstoneMentionLinkTx(ctx, tx, before, *before.ResolvedRecordID)
		if err != nil {
			return mentionMutationResult{}, err
		}
		if removeOldLink {
			existingLink, err := s.ports.links.GetActiveLinkTx(ctx, tx, before.IncidentID, before.SourceRecordID, *before.ResolvedRecordID, linkType)
			switch {
			case errors.Is(err, errRecordLinkNotFound):
			case err != nil:
				return mentionMutationResult{}, err
			default:
				tombstoned, err := s.ports.links.TombstoneLinkTx(ctx, tx, existingLink.RecordLinkID, actorUserID, now.UTC())
				if err != nil {
					return mentionMutationResult{}, err
				}
				tombstonedLink = &tombstoned
			}
		}
	}
	if after.ResolvedRecordID != nil {
		link, inserted, err := s.ports.links.UpsertLinkTx(ctx, tx, after.IncidentID, after.SourceRecordID, *after.ResolvedRecordID, linkType, "manual", nil, actorUserID, now.UTC())
		if err != nil {
			return mentionMutationResult{}, err
		}
		activeLink = &link
		if inserted {
			createdLink = &link
		}
	}
	if err := s.refreshMentionEntityRowsTx(ctx, tx, before, after); err != nil {
		return mentionMutationResult{}, err
	}

	return mentionMutationResult{
		Before:         before,
		After:          after,
		ActiveLink:     activeLink,
		CreatedLink:    createdLink,
		TombstonedLink: tombstonedLink,
	}, nil
}

func (s *Store) refreshMentionEntityRowsTx(ctx context.Context, tx pgx.Tx, before mentionActionRecord, after mentionActionRecord) error {
	seen := map[uuid.UUID]struct{}{}
	for _, recordID := range []*uuid.UUID{before.ResolvedRecordID, after.ResolvedRecordID} {
		if recordID == nil {
			continue
		}
		if _, ok := seen[*recordID]; ok {
			continue
		}
		seen[*recordID] = struct{}{}
		if err := s.ports.projections.RefreshEntityRowTx(ctx, tx, *recordID, after.EntityType); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) mentionEntityInvalidationsTx(ctx context.Context, tx pgx.Tx, outcome mentionMutationResult) ([]MentionEntityInvalidation, error) {
	viewSchemaID, changedFieldKey, ok := mentionEntityInvalidationSurface(outcome.After.EntityType)
	if !ok {
		return nil, nil
	}

	seen := make(map[uuid.UUID]struct{}, 2)
	recordIDs := make([]uuid.UUID, 0, 2)
	for _, recordID := range []*uuid.UUID{outcome.Before.ResolvedRecordID, outcome.After.ResolvedRecordID} {
		if recordID == nil {
			continue
		}
		if _, ok := seen[*recordID]; ok {
			continue
		}
		seen[*recordID] = struct{}{}
		recordIDs = append(recordIDs, *recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})

	invalidations := make([]MentionEntityInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		rowVersion, err := s.ports.records.LoadRowVersionTx(ctx, tx, recordID)
		if err != nil {
			return nil, fmt.Errorf("load mention entity invalidation row version: %w", err)
		}
		invalidations = append(invalidations, MentionEntityInvalidation{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ViewSchemaID:     viewSchemaID,
			ChangedFieldKeys: []string{changedFieldKey},
		})
	}
	return invalidations, nil
}

func mentionEntityInvalidationSurface(entityType string) (string, string, bool) {
	switch entityType {
	case "host":
		return entitycontract.HostsViewSchemaID, "host.linked_event_count", true
	case "identity":
		return entitycontract.IdentitiesViewSchemaID, "identity.linked_event_count", true
	default:
		return "", "", false
	}
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
	if rows.Next() {
		var record mentionTargetRecord
		if err := rows.Scan(&record.EntityType, &record.IncidentID, &record.RecordID); err != nil {
			return nil, fmt.Errorf("scan mention resolved target: %w", err)
		}
		target = &record
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

func buildMentionActionPayload(incidentID uuid.UUID, mention mentionActionRecord, sourceRecordID uuid.UUID, sourceRecordRowVersion int64, changeSetID uuid.UUID, activeLink *recordLink) map[string]any {
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

func buildLinkMutationValue(link recordLink, deletedAt *time.Time) map[string]any {
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

func mentionChangedFieldKeys(sourceFieldKey string) []string {
	keys := []string{sourceFieldKey, "timeline.has_unresolved_mentions"}
	slices.Sort(keys)
	return keys
}

func mentionIdempotencyScope(mentionID uuid.UUID) string {
	return mentionID.String()
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
