package assessments

import (
	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	return incidentbundle.NewSourcePort()
}
