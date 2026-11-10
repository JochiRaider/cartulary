package projectionprovider

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityfacts "github.com/JochiRaider/cartulary/internal/modules/entities/timelinefacts"
	evidencefacts "github.com/JochiRaider/cartulary/internal/modules/evidence/timelinefacts"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/collectionfacts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

// NewSource constructs the canonical Timeline projection source from
// Timeline's declared source-authority readers.
func NewSource(linkReader collectionfacts.LinkReader) *workbookprojection.Source {
	return workbookprojection.NewSource(
		timelineEnvelopeReader{store: recordenvelope.Reader{}},
		collectionfacts.New(
			entityfacts.Reader{},
			linkReader,
			evidencefacts.TimelineFactReader{},
		),
	)
}

type timelineEnvelopeReader struct {
	store recordenvelope.Reader
}

func (reader timelineEnvelopeReader) LoadEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (sourcerepository.Envelope, error) {
	envelope, err := reader.store.LoadEnvelopeTx(ctx, tx, recordID)
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return sourcerepository.Envelope{}, sourcerepository.ErrEnvelopeNotFound
	}
	return projectionTimelineEnvelope(envelope), err
}

func (reader timelineEnvelopeReader) LoadEnvelopesTx(
	ctx context.Context,
	tx pgx.Tx,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]sourcerepository.Envelope, error) {
	envelopes, err := reader.store.LoadEnvelopesTx(ctx, tx, recordIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]sourcerepository.Envelope, len(envelopes))
	for recordID, envelope := range envelopes {
		result[recordID] = projectionTimelineEnvelope(envelope)
	}
	return result, nil
}

func projectionTimelineEnvelope(envelope recordenvelope.Envelope) sourcerepository.Envelope {
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
