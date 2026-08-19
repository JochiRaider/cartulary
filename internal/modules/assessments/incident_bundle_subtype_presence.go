package assessments

import (
	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

func IncidentBundleSubtypeContribution() subtypepresence.Contribution {
	return incidentbundle.SubtypeContribution()
}
