package artifacts

import (
	artifactprojectionprovider "github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/providers/projection"
	artifactreportingprovider "github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/providers/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

// NewProjectionContribution constructs the Artifacts source contribution while
// leaving projection storage and coordination in Projections.
func NewProjectionContribution() (workbookprojection.Contribution, error) {
	return workbookprojection.NewContribution(artifactprojectionprovider.NewSource())
}

// NewReportingContribution constructs the Artifacts field provider while
// leaving report composition and export coordination in Reporting.
func NewReportingContribution(reader workbookprojection.Reader) (exportprovider.FieldProvider, error) {
	return artifactreportingprovider.New(reader)
}
