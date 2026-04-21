package entities

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrInvalidMentionResolution = errors.New("entities: invalid mention resolution")

func (s *Store) ResolveOrCreateFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) (MentionResolutionResult, error) {
	mention, err := loadMentionActionRecordTx(ctx, tx, mentionID)
	if err != nil {
		return MentionResolutionResult{}, err
	}
	if mention.SourceRecordID != sourceRecordID || mention.SourceFieldKey != fieldKey {
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}
	if resolvedRecordID == nil && mention.ResolutionStatus != "unresolved" {
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	var (
		result           MentionResolutionResult
		resolutionMethod = resolutionMethodPointer("created_from_mention")
		validatedTarget  *mentionTargetRecord
	)
	switch mention.EntityType {
	case "host":
		if resolvedRecordID != nil {
			target, err := validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *resolvedRecordID)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			validatedTarget = target
			result = MentionResolutionResult{
				EntityType: "host",
				RecordID:   target.RecordID,
			}
			resolutionMethod = resolutionMethodPointer("explicit_resolve_route")
		} else {
			record, beforeRow, operationKind, _, err := s.upsertHostWithInputTx(ctx, tx, actor, mention.IncidentID, hostInputFromMention(mentionRecordFromAction(mention)), now)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			if err := upsertHostProjectionTx(ctx, tx, record); err != nil {
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
			validatedTarget = &mentionTargetRecord{
				RecordID:   record.RecordID,
				IncidentID: mention.IncidentID,
				EntityType: mention.EntityType,
			}
		}
	case "identity":
		if resolvedRecordID != nil {
			target, err := validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *resolvedRecordID)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			validatedTarget = target
			result = MentionResolutionResult{
				EntityType: "identity",
				RecordID:   target.RecordID,
			}
			resolutionMethod = resolutionMethodPointer("explicit_resolve_route")
		} else {
			record, beforeRow, operationKind, _, err := s.upsertIdentityWithInputTx(ctx, tx, actor, mention.IncidentID, identityInputFromMention(mentionRecordFromAction(mention)), now)
			if err != nil {
				return MentionResolutionResult{}, err
			}
			if err := upsertIdentityProjectionTx(ctx, tx, record); err != nil {
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
			validatedTarget = &mentionTargetRecord{
				RecordID:   record.RecordID,
				IncidentID: mention.IncidentID,
				EntityType: mention.EntityType,
			}
		}
	default:
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	if _, err := s.applyMentionActionTx(ctx, tx, actor.ID, mention, "resolve_item", validatedTarget, resolutionMethod, now.UTC()); err != nil {
		return MentionResolutionResult{}, err
	}
	return result, nil
}

func mentionRecordFromAction(record mentionActionRecord) mentionRecord {
	return mentionRecord{
		EntityMentionID:  record.EntityMentionID,
		SourceRecordID:   record.SourceRecordID,
		IncidentID:       record.IncidentID,
		EntityType:       record.EntityType,
		SourceFieldKey:   record.SourceFieldKey,
		RawText:          record.RawText,
		ResolutionStatus: record.ResolutionStatus,
	}
}
