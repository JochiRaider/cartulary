package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	indicatorbundle "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/sourcestate"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

// IncidentBundleContribution is the complete Indicator-owned portability
// boundary. The Incident Bundles coordinator receives generic source-port and
// subtype-presence contracts, never Indicator implementation details.
type IncidentBundleContribution struct {
	SourcePort      sourceport.Port
	SubtypePresence subtypepresence.Contribution
}

func NewIncidentBundleContribution() (IncidentBundleContribution, error) {
	catalog, err := sourcestate.Load()
	if err != nil {
		return IncidentBundleContribution{}, err
	}
	paths := make([]sourceport.Path, 0, len(catalog.PortabilityDescriptors()))
	for _, descriptor := range catalog.PortabilityDescriptors() {
		paths = append(paths, sourceport.Path{
			LogicalPath:               descriptor.LogicalPath,
			ContentRole:               descriptor.ContentRole,
			SchemaID:                  descriptor.SchemaID,
			Versions:                  append([]int(nil), descriptor.Versions...),
			StableIdentity:            append([]string(nil), descriptor.StableIdentity...),
			StableIdentityInvariantID: descriptor.StableIdentityInvariantID,
		})
	}
	return IncidentBundleContribution{
		SourcePort:      indicatorbundle.NewSourcePort(paths),
		SubtypePresence: indicatorbundle.SubtypeContribution(),
	}, nil
}
