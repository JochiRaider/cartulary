package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/incidentbundle"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	return incidentbundle.NewSourcePort()
}

func IncidentBundleSubtypeContribution() subtypepresence.Contribution {
	return incidentbundle.SubtypeContribution()
}
