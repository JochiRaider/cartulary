package timelineassembly

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type mentionAdapter struct {
	store *mentions.Store
}

func (a mentionAdapter) TransitionStatus(field, from, action string) (string, error) {
	return a.store.TransitionStatus(field, from, action)
}

func (a mentionAdapter) LoadCollectionSourceFactsTx(ctx context.Context, tx pgx.Tx, sourceID uuid.UUID, field string) ([]timeline.MentionSourceFact, error) {
	facts, err := a.store.LoadCollectionSourceFactsTx(ctx, tx, sourceID, field)
	if errors.Is(err, mentions.ErrInvalidMentionResolution) {
		return nil, &timeline.InvalidCollectionActionError{FieldKey: field}
	}
	if err != nil {
		return nil, err
	}
	result := make([]timeline.MentionSourceFact, len(facts))
	for i, fact := range facts {
		result[i] = timeline.MentionSourceFact(fact)
	}
	return result, nil
}

func (a mentionAdapter) ResolveExistingFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) ([]timeline.RecordLinkMutation, error) {
	result, err := a.store.ResolveExistingFromMentionTx(ctx, tx, actor, sourceRecordID, fieldKey, mentionID, resolvedRecordID, now)
	if errors.Is(err, mentions.ErrResolvedRecordNotFound) {
		return nil, timeline.ErrResolvedRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return timelineMentionLinkMutations(result.LinkMutations), nil
}

func (a mentionAdapter) ApplyMentionLifecycleTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, sourceFieldKey string, mentionID uuid.UUID, action string, resolvedRecordID *uuid.UUID, now time.Time) ([]timeline.RecordLinkMutation, error) {
	mutations, err := a.store.ApplyMentionLifecycleTx(ctx, tx, actor, sourceRecordID, sourceFieldKey, mentionID, action, resolvedRecordID, now)
	if errors.Is(err, mentions.ErrResolvedRecordNotFound) {
		return nil, timeline.ErrResolvedRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return timelineMentionLinkMutations(mutations), nil
}

func timelineMentionLinkMutations(mutations []mentions.LinkMutation) []timeline.RecordLinkMutation {
	converted := make([]timeline.RecordLinkMutation, 0, len(mutations))
	for _, mutation := range mutations {
		converted = append(converted, timeline.RecordLinkMutation{
			RecordLinkID: mutation.RecordLinkID,
			Operation:    mutation.Operation,
			BeforeValue:  cloneLinkMutationMap(mutation.BeforeValue),
			AfterValue:   cloneLinkMutationMap(mutation.AfterValue),
		})
	}
	return converted
}

func (a mentionAdapter) NextOrdinalTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string) (int, error) {
	return a.store.NextOrdinalTx(ctx, tx, recordID, fieldKey)
}

func (a mentionAdapter) InsertTx(ctx context.Context, tx pgx.Tx, params timeline.MentionCreateParams) error {
	return a.store.InsertTx(ctx, tx, mentions.CreateParams(params))
}

type entityAdapter struct {
	store *hostidentity.SourceFacts
}

func (a entityAdapter) ListEligibleAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]timeline.EntityAlias, error) {
	aliases, err := a.store.ListEligibleAliasesTx(ctx, tx, incidentID, entityType)
	if err != nil {
		return nil, err
	}
	result := make([]timeline.EntityAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, timeline.EntityAlias(alias))
	}
	return result, nil
}

func (a entityAdapter) InspectResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	err := a.store.InspectResolvedTargetTx(ctx, tx, incidentID, entityType, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return timeline.ErrResolvedRecordNotFound
	}
	return err
}

func (a entityAdapter) AdmitResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	err := a.store.AdmitResolvedTargetTx(ctx, tx, incidentID, entityType, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return timeline.ErrResolvedRecordNotFound
	}
	return err
}

type evidenceAdapter struct {
	attachments evidence.TimelineAttachmentContribution
}

func (a evidenceAdapter) ValidateTimelineAttachmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) error {
	err := a.attachments.ValidateTimelineAttachmentsTx(ctx, tx, incidentID, recordIDs)
	if errors.Is(err, evidence.ErrEvidenceNotFound) {
		return &links.CollectionValidationError{
			Field:      "timeline.attached_evidence_ids",
			ReasonCode: "invalid_value",
		}
	}
	return err
}

func (a evidenceAdapter) RefreshTimelineAttachmentProjectionsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	return a.attachments.RefreshTimelineAttachmentProjectionsTx(ctx, tx, recordIDs)
}
