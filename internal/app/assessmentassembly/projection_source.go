package assessmentassembly

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

func NewProjectionContribution() (assessmentprojection.Contribution, error) {
	recordStore := records.NewStore()
	return assessments.NewProjectionContribution(assessments.ProjectionContributionDependencies{
		Envelopes: assessmentEnvelopeAdapter{records: recordStore},
		Support: assessmentSupportAdapter{
			links:   links.FactReader{},
			records: recordStore,
		},
	})
}

type assessmentEnvelopeAdapter struct {
	records *records.Store
}

func (adapter assessmentEnvelopeAdapter) LoadAssessmentProjectionEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (assessmentprojection.Envelope, bool, error) {
	envelope, err := adapter.records.LoadEnvelopeTx(ctx, tx, recordID, false)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
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

type assessmentSupportAdapter struct {
	links   links.FactReader
	records *records.Store
}

func (adapter assessmentSupportAdapter) LoadAssessmentProjectionSupportFactsTx(
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
		if fact.SrcRecordID == assessmentID && fact.LinkType == links.LinkTypeSupportedBy {
			targetRecordIDs = append(targetRecordIDs, fact.DstRecordID)
		}
	}
	envelopes, err := adapter.records.LoadEnvelopesTx(ctx, tx, targetRecordIDs, false)
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
