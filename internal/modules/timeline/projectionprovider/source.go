package projectionprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityfacts "github.com/JochiRaider/cartulary/internal/modules/entities/timelinefacts"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

// NewSource constructs the canonical Timeline projection source from
// Timeline's declared source-authority readers.
func NewSource() *workbookprojection.Source {
	return workbookprojection.NewSource(
		timelineEnvelopeReader{store: records.NewStore()},
		timelineCollectionFacts{},
	)
}

type timelineEnvelopeReader struct {
	store *records.Store
}

func (reader timelineEnvelopeReader) LoadEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	lock bool,
) (sourcerepository.Envelope, error) {
	envelope, err := reader.store.LoadEnvelopeTx(ctx, tx, recordID, lock)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return sourcerepository.Envelope{}, sourcerepository.ErrEnvelopeNotFound
	}
	return projectionTimelineEnvelope(envelope), err
}

func (reader timelineEnvelopeReader) LoadEnvelopesTx(
	ctx context.Context,
	tx pgx.Tx,
	recordIDs []uuid.UUID,
	lock bool,
) (map[uuid.UUID]sourcerepository.Envelope, error) {
	envelopes, err := reader.store.LoadEnvelopesTx(ctx, tx, recordIDs, lock)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]sourcerepository.Envelope, len(envelopes))
	for recordID, envelope := range envelopes {
		result[recordID] = projectionTimelineEnvelope(envelope)
	}
	return result, nil
}

func projectionTimelineEnvelope(envelope records.Envelope) sourcerepository.Envelope {
	return sourcerepository.Envelope{
		RecordID:        envelope.RecordID,
		IncidentID:      envelope.IncidentID,
		RecordType:      envelope.RecordType,
		RowVersion:      envelope.RowVersion,
		CreatedByUserID: envelope.CreatedByUserID,
		CreatedAt:       envelope.CreatedAt,
		UpdatedByUserID: envelope.UpdatedByUserID,
		UpdatedAt:       envelope.UpdatedAt,
		DeletedAt:       envelope.DeletedAt,
	}
}

type timelineCollectionFacts struct{}

func (timelineCollectionFacts) LoadTimelineCollectionFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (workbookprojection.CollectionFacts, error) {
	mentions, err := (entityfacts.Reader{}).LoadMentionsTx(ctx, tx, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline entity facts: %w", err)
	}
	linkFacts, err := (links.TimelineFactReader{}).LoadTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline link facts: %w", err)
	}
	evidenceFacts, err := (evidence.TimelineFactReader{}).LoadTx(ctx, tx, incidentID, linkFacts.AttachedEvidenceIDs)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline evidence facts: %w", err)
	}
	attachedEvidence := make([]workbookprojection.EvidenceFact, len(evidenceFacts))
	for index, fact := range evidenceFacts {
		attachedEvidence[index] = workbookprojection.EvidenceFact{
			RecordID:       fact.RecordID,
			Title:          fact.Title,
			LifecycleState: fact.LifecycleState,
			UploadState:    fact.UploadState,
		}
	}
	return workbookprojection.CollectionFacts{
		Mentions:            mentions,
		ResolvedLinks:       linkFacts.ResolvedLinks,
		Tags:                linkFacts.Tags,
		AttachedEvidence:    attachedEvidence,
		ReplacementRecordID: linkFacts.ReplacementRecordID,
	}, nil
}
