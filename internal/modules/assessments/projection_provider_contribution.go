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
	if dependencies.Envelopes == nil {
		return workbookprojection.Contribution{}, errors.New(
			"construct assessment projection contribution: envelope reader is required",
		)
	}
	if dependencies.Support == nil {
		return workbookprojection.Contribution{}, errors.New(
			"construct assessment projection contribution: support fact reader is required",
		)
	}
	return workbookprojection.NewContribution(assessmentprovider.NewSource(
		dependencies.Envelopes,
		dependencies.Support,
	))
}
