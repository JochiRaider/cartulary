package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

type IncidentBundleContribution struct {
	SourcePort      sourceport.Port
	SubtypePresence subtypepresence.Contribution
}

func NewIncidentBundleContribution() IncidentBundleContribution {
	return IncidentBundleContribution{
		SourcePort:      newIncidentBundleSourcePort(),
		SubtypePresence: incidentBundleSubtypeContribution(),
	}
}
