package projectionprovider

import (
	internalprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/projection"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

func NewContribution() (workbookprojection.Contribution, error) {
	return workbookprojection.NewRuntimeContribution(
		internalprojection.NewTaskRequestSource(),
		internalprojection.NewDecisionSource(),
	)
}
