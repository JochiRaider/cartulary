package mentions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrInvalidMentionResolution = newInvalidMutationTargetError("entities: invalid mention resolution")

type MentionResolutionResult struct {
	EntityType    string
	RecordID      uuid.UUID
	OperationKind string
	BeforeRow     map[string]any
	AfterRow      map[string]any
	LinkMutations []LinkMutation
}

func (s *Store) ResolveExistingFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) (MentionResolutionResult, error) {
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

	outcome, err := s.applyMentionActionTx(ctx, tx, actor.ID, mention, "resolve_item", validatedTarget, resolutionMethod, now.UTC())
	if err != nil {
		return MentionResolutionResult{}, err
	}
	result.LinkMutations = append([]LinkMutation(nil), outcome.LinkMutations...)
	return result, nil
}
