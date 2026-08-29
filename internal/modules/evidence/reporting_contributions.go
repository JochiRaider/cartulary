package evidence

import (
	evidencereporting "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func NewReportingFieldContribution() exportprovider.FieldProvider {
	return evidencereporting.NewFieldContribution()
}

func NewReportingLogicalTargetContribution() exportprovider.LogicalSupportTargetProvider {
	return evidencereporting.NewLogicalTargetContribution()
}
