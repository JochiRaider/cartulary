package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/incidentbundle"
)

type IncidentBundleContribution struct {
	SourcePort      sourceport.Port
	SubtypePresence subtypepresence.Contribution
}

func NewIncidentBundleContribution() IncidentBundleContribution {
	return IncidentBundleContribution{
		SourcePort:      incidentbundle.NewSourcePort(),
		SubtypePresence: incidentbundle.SubtypeContribution(),
	}
}
