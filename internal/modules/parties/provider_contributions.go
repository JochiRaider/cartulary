package parties

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	partyincidentbundle "github.com/JochiRaider/cartulary/internal/modules/parties/internal/providers/incidentbundle"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/internal/providers/projection"
	partyreporting "github.com/JochiRaider/cartulary/internal/modules/parties/internal/providers/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

// NewProjectionContribution constructs the Party source contribution while
// leaving projection storage and coordination in Projections.
func NewProjectionContribution() (workbookprojection.Contribution, error) {
	return partyprojection.NewContribution()
}

// NewReportingContribution constructs the Party source-fact provider while
// leaving export coordination and report semantics in Reporting.
func NewReportingContribution() exportprovider.FieldProvider {
	return partyreporting.New()
}

// NewIncidentBundleContribution constructs the Party source port while
// leaving portable operation coordination in Incident Bundles.
func NewIncidentBundleContribution() sourceport.Port {
	return partyincidentbundle.NewContribution()
}

func IncidentBundleSubtypeContribution() subtypepresence.Contribution {
	return partyincidentbundle.SubtypeContribution()
}
