package projectionfacts

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	linkvocabulary "github.com/JochiRaider/cartulary/internal/modules/links/vocabulary"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
)

func NewEnvelopeReader() EnvelopeAdapter { return EnvelopeAdapter{records: recordenvelope.Reader{}} }
func NewSupportReader() SupportAdapter {
	return SupportAdapter{records: recordenvelope.Reader{}, links: linkfacts.FactReader{}}
}

type EnvelopeAdapter struct {
	records recordenvelope.Reader
}

func (adapter EnvelopeAdapter) LoadAssessmentProjectionEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (assessmentprojection.Envelope, bool, error) {
	envelope, err := adapter.records.LoadEnvelopeTx(ctx, tx, recordID)
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return assessmentprojection.Envelope{}, false, nil
	}
	if err != nil {
		return assessmentprojection.Envelope{}, false, err
	}
	return assessmentprojection.Envelope{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		RowVersion: envelope.RowVersion,
		DeletedAt:  envelope.DeletedAt,
	}, true, nil
}

type SupportAdapter struct {
	links   linkfacts.FactReader
	records recordenvelope.Reader
}

func (adapter SupportAdapter) LoadAssessmentProjectionSupportFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	assessmentID uuid.UUID,
) (assessmentprojection.SupportFacts, error) {
	linkFacts, err := adapter.links.LoadRecordTx(ctx, tx, incidentID, assessmentID)
	if err != nil {
		return assessmentprojection.SupportFacts{}, err
	}
	targetRecordIDs := make([]uuid.UUID, 0)
	for _, fact := range linkFacts.RecordLinks {
		if fact.SrcRecordID == assessmentID && fact.LinkType == linkvocabulary.LinkTypeSupportedBy {
			targetRecordIDs = append(targetRecordIDs, fact.DstRecordID)
		}
	}
	envelopes, err := adapter.records.LoadEnvelopesTx(ctx, tx, targetRecordIDs)
	if err != nil {
		return assessmentprojection.SupportFacts{}, err
	}
	activeTargets := 0
	for _, targetID := range targetRecordIDs {
		envelope, ok := envelopes[targetID]
		if ok && envelope.IncidentID == incidentID && envelope.DeletedAt == nil {
			activeTargets++
		}
	}
	return assessmentprojection.SupportFacts{ActiveTargetCount: activeTargets}, nil
}
