package importassembly

import (
	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
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
	entityProjections entityprojection.Writer,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	entityStore := hostidentity.NewStore(pool, appender, nil, entityProjections)
	return assessments.NewImportCreateFacade(
		targetViewSchemaID,
		facadeID,
		assessments.ImportCreateDependencies{
			Subjects:    assessmentassembly.NewSubjectValidator(pool, entityStore),
			Assessors:   assessmentassembly.NewAssessorValidator(pool),
			Records:     assessmentassembly.NewRecordEnvelopeCreator(pool),
			Revisions:   appender,
			Projections: assessmentassembly.NewProjectionPort(projectionRows),
		},
	)
}
