package entities

import (
	"context"
	"errors"
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
	if resolvedRecordID == nil {
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	var (
		result           MentionResolutionResult
		resolutionMethod = resolutionMethodPointer("explicit_resolve_route")
		validatedTarget  *mentionTargetRecord
	)
	switch mention.EntityType {
	case "host":
		target, err := validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *resolvedRecordID)
		if err != nil {
			return MentionResolutionResult{}, err
		}
		validatedTarget = target
		result = MentionResolutionResult{
			EntityType: "host",
			RecordID:   target.RecordID,
		}
	case "identity":
		target, err := validateMentionResolvedTargetTx(ctx, tx, actor.ID, mention.IncidentID, mention.EntityType, *resolvedRecordID)
		if err != nil {
			return MentionResolutionResult{}, err
		}
		validatedTarget = target
		result = MentionResolutionResult{
			EntityType: "identity",
			RecordID:   target.RecordID,
		}
	default:
		return MentionResolutionResult{}, ErrInvalidMentionResolution
	}

	if _, err := s.applyMentionActionTx(ctx, tx, actor.ID, mention, "resolve_item", validatedTarget, resolutionMethod, now.UTC()); err != nil {
		return MentionResolutionResult{}, err
	}
	return result, nil
}
