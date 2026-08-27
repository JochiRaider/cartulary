package assessments

import (
	"errors"

	assessmentprovider "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/providers/projection"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

type ProjectionContributionDependencies struct {
	Envelopes workbookprojection.EnvelopeReader
	Support   workbookprojection.SupportFactReader
}

func NewProjectionContribution(
	dependencies ProjectionContributionDependencies,
) (workbookprojection.Contribution, error) {
	if isNilDependency(dependencies.Envelopes) {
		return workbookprojection.Contribution{}, errors.New(
			"construct assessment projection contribution: envelope reader is required",
		)
	}
	if isNilDependency(dependencies.Support) {
		return workbookprojection.Contribution{}, errors.New(
			"construct assessment projection contribution: support fact reader is required",
		)
	}
	return workbookprojection.NewContribution(assessmentprovider.NewSource(
		dependencies.Envelopes,
		dependencies.Support,
	))
}
