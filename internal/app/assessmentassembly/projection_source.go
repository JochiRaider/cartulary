package assessmentassembly

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprovider "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

func NewProjectionContribution() (assessmentprojection.Contribution, error) {
	recordStore := records.NewStore()
	source := assessmentprovider.NewSource(
		assessmentEnvelopeAdapter{records: recordStore},
		assessmentSupportAdapter{
			links:   links.AssessmentFactReader{},
			records: recordStore,
		},
	)
	return assessmentprojection.NewContribution(source)
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
		RecordID:   envelope.RecordID,
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		RowVersion: envelope.RowVersion,
		DeletedAt:  envelope.DeletedAt,
	}, true, nil
}

type assessmentSupportAdapter struct {
	links   links.AssessmentFactReader
	records *records.Store
}

func (adapter assessmentSupportAdapter) LoadAssessmentProjectionSupportFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	assessmentID uuid.UUID,
) (assessmentprojection.SupportFacts, error) {
	linkFacts, err := adapter.links.LoadSupportFactsTx(ctx, tx, incidentID, assessmentID)
	if err != nil {
		return assessmentprojection.SupportFacts{}, err
	}
	envelopes, err := adapter.records.LoadEnvelopesTx(ctx, tx, linkFacts.TargetRecordIDs, false)
	if err != nil {
		return assessmentprojection.SupportFacts{}, err
	}
	activeTargets := 0
	for _, targetID := range linkFacts.TargetRecordIDs {
		envelope, ok := envelopes[targetID]
		if ok && envelope.IncidentID == incidentID && envelope.DeletedAt == nil {
			activeTargets++
		}
	}
	return assessmentprojection.SupportFacts{ActiveTargetCount: activeTargets}, nil
}
