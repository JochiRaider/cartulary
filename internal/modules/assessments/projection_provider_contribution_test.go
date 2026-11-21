package assessments_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

func TestAssessmentProjectionContributionConstruction_Unit(t *testing.T) {
	t.Parallel()
	for name, dependencies := range map[string]assessments.ProjectionContributionDependencies{
		"missing envelopes": {Support: assessmentProjectionSupportStub{}},
		"missing support":   {Envelopes: assessmentProjectionEnvelopeStub{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := assessments.NewProjectionContribution(dependencies); err == nil {
				t.Fatal("incomplete assessment projection contribution unexpectedly constructed")
			}
		})
	}
	contribution, err := assessments.NewProjectionContribution(
		assessments.ProjectionContributionDependencies{
			Envelopes: assessmentProjectionEnvelopeStub{},
			Support:   assessmentProjectionSupportStub{},
		},
	)
	if err != nil {
		t.Fatalf("construct assessment projection contribution: %v", err)
	}
	descriptor := contribution.ProjectionContribution().Descriptors()[0]
	if descriptor.ProviderID != "assessment" || descriptor.SourceOwnerModule != "assessments" {
		t.Fatalf("assessment projection descriptor drifted: %#v", descriptor)
	}
}

type assessmentProjectionEnvelopeStub struct{}

func (assessmentProjectionEnvelopeStub) LoadAssessmentProjectionEnvelopesTx(
	context.Context,
	pgx.Tx,
	[]uuid.UUID,
) (map[uuid.UUID]workbookprojection.Envelope, error) {
	return map[uuid.UUID]workbookprojection.Envelope{}, nil
}

type assessmentProjectionSupportStub struct{}

func (assessmentProjectionSupportStub) LoadAssessmentProjectionSupportFactsTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
	[]uuid.UUID,
) (map[uuid.UUID]workbookprojection.SupportFacts, error) {
	return map[uuid.UUID]workbookprojection.SupportFacts{}, nil
}
