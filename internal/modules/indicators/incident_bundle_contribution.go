package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	indicatorbundle "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

// IncidentBundleContribution is the complete Indicator-owned portability
// boundary. The Incident Bundles coordinator receives generic source-port and
// subtype-presence contracts, never Indicator implementation details.
type IncidentBundleContribution struct {
	SourcePort      sourceport.Port
	SubtypePresence subtypepresence.Contribution
}

func NewIncidentBundleContribution() IncidentBundleContribution {
	return IncidentBundleContribution{
		SourcePort:      indicatorbundle.NewSourcePort(),
		SubtypePresence: indicatorbundle.SubtypeContribution(),
	}
}
