package importassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func newAssessmentImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	pool postgres.DB,
	projectionRows assessmentprojection.Rows,
	appender *revisions.Appender,
	publications collaboration.RecordChangedAppender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	subjects, err := assessmentassembly.NewSubjectValidator(pool, hostidentity.NewSourceFacts())
	if err != nil {
		return nil, fmt.Errorf("compose assessment import facade: %w", err)
	}
	assessors, err := assessmentassembly.NewAssessorValidator(pool)
	if err != nil {
		return nil, fmt.Errorf("compose assessment import facade: %w", err)
	}
	records, err := assessmentassembly.NewRecordEnvelopeCreator(pool)
	if err != nil {
		return nil, fmt.Errorf("compose assessment import facade: %w", err)
	}
	projections, err := assessmentassembly.NewProjectionPort(projectionRows)
	if err != nil {
		return nil, fmt.Errorf("compose assessment import facade: %w", err)
	}
	return assessments.NewImportCreateFacade(
		targetViewSchemaID,
		facadeID,
		assessments.ImportCreateDependencies{
			Subjects:      subjects,
			Assessors:     assessors,
			Records:       records,
			Revisions:     appender,
			Projections:   projections,
			Collaboration: publications,
		},
	)
}
