package projectionprovider

import "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"

func NewContribution() (workbookprojection.Contribution, error) {
	return workbookprojection.NewRuntimeContribution(NewSource())
}
