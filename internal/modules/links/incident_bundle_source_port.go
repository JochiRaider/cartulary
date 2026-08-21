package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/incidentbundle"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	return incidentbundle.NewSourcePort()
}
