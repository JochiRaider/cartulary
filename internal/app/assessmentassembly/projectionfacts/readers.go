package projectionfacts

import (
	"context"

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

func (adapter EnvelopeAdapter) LoadAssessmentProjectionEnvelopesTx(
	ctx context.Context,
	tx pgx.Tx,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]assessmentprojection.Envelope, error) {
	envelopes, err := adapter.records.LoadEnvelopesTx(ctx, tx, recordIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]assessmentprojection.Envelope, len(envelopes))
	for recordID, envelope := range envelopes {
		result[recordID] = assessmentprojection.Envelope{
			IncidentID: envelope.IncidentID,
			RecordType: envelope.RecordType,
			RowVersion: envelope.RowVersion,
			DeletedAt:  envelope.DeletedAt,
		}
	}
	return result, nil
}

type SupportAdapter struct {
	links   linkfacts.FactReader
	records recordenvelope.Reader
}

func (adapter SupportAdapter) LoadAssessmentProjectionSupportFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	assessmentIDs []uuid.UUID,
) (map[uuid.UUID]assessmentprojection.SupportFacts, error) {
	result := make(map[uuid.UUID]assessmentprojection.SupportFacts, len(assessmentIDs))
	if len(assessmentIDs) == 0 {
		return result, nil
	}
	linkFacts, err := adapter.links.LoadRecordsTx(ctx, tx, incidentID, assessmentIDs)
	if err != nil {
		return nil, err
	}
	requested := make(map[uuid.UUID]bool, len(assessmentIDs))
	for _, assessmentID := range assessmentIDs {
		requested[assessmentID] = true
	}
	targetsByAssessment := make(map[uuid.UUID][]uuid.UUID)
	targetRecordIDs := make([]uuid.UUID, 0)
	for _, fact := range linkFacts.RecordLinks {
		if requested[fact.SrcRecordID] && fact.LinkType == linkvocabulary.LinkTypeSupportedBy {
			targetsByAssessment[fact.SrcRecordID] = append(targetsByAssessment[fact.SrcRecordID], fact.DstRecordID)
			targetRecordIDs = append(targetRecordIDs, fact.DstRecordID)
		}
	}
	envelopes, err := adapter.records.LoadEnvelopesTx(ctx, tx, targetRecordIDs)
	if err != nil {
		return nil, err
	}
	for _, assessmentID := range assessmentIDs {
		activeTargets := 0
		for _, targetID := range targetsByAssessment[assessmentID] {
			envelope, ok := envelopes[targetID]
			if ok && envelope.IncidentID == incidentID && envelope.DeletedAt == nil {
				activeTargets++
			}
		}
		result[assessmentID] = assessmentprojection.SupportFacts{ActiveTargetCount: activeTargets}
	}
	return result, nil
}
