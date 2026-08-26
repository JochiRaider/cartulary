package tasksdecisions

import (
	internalprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/projection"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

func NewProjectionContribution() (projectioncontract.Contribution, error) {
	return projectioncontract.NewContribution(
		internalprojection.NewTaskRequestSource(),
		internalprojection.NewDecisionSource(),
	)
}
