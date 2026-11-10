package assessmentassembly

import (
	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly/projectionfacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

func NewProjectionContribution() (assessmentprojection.Contribution, error) {
	return assessments.NewProjectionContribution(assessments.ProjectionContributionDependencies{Envelopes: projectionfacts.NewEnvelopeReader(), Support: projectionfacts.NewSupportReader()})
}
