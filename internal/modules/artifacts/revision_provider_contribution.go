package artifacts

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/providers/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func NewRevisionContribution() (revisions.ProviderContribution, error) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return revisions.ProviderContribution{}, err
	}
	routes := make([]revisions.RecordViewRouteContribution, 0, 8)
	for _, surface := range catalog.Surfaces() {
		routes = append(routes, recordViewRoute(surface))
	}
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerArtifacts,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerArtifacts,
			RecordType:          "artifact",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.artifact.v1",
			DeleteRestoreSource: deleterestore.NewSource(catalog),
			RowRollbackProvider: rollback.NewProvider(),
			RecordViewRoutes:    routes,
		}},
	}, nil
}

func recordViewRoute(surface sourcecatalog.Surface) revisions.RecordViewRouteContribution {
	return revisions.RecordViewRouteContribution{
		ContributionID: "artifacts." + viewSchemaKey(surface.ViewSchemaID),
		Variant: &revisions.RecordVariant{
			Kind:  "artifact_type",
			Value: surface.ArtifactType,
		},
		ViewSchemaIDs: []string{surface.ViewSchemaID},
	}
}

func viewSchemaKey(viewSchemaID string) string {
	const (
		prefix = "cartulary.view."
		suffix = ".v1"
	)
	return viewSchemaID[len(prefix) : len(viewSchemaID)-len(suffix)]
}
