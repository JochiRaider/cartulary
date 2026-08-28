package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/sourcestate"
)

func NewIncidentBundleSourcePort() (sourceport.Port, error) {
	return sourcestate.NewSourcePort()
}
